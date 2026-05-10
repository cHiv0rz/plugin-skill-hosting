import { describe, it, expect, beforeEach, vi } from 'vitest'
import { nextTick, ref } from 'vue'

vi.mock('../api', () => ({
  api: {
    listSkillFiles: vi.fn(),
    getSkillFile: vi.fn(),
    putSkillFile: vi.fn(),
    deleteSkillFile: vi.fn(),
  },
  errMsg: (e: unknown, fallback = 'something went wrong') =>
    e instanceof Error ? e.message : fallback,
}))

const confirmMock = vi.fn()
vi.mock('./useConfirm', () => ({
  useConfirm: () => ({ confirm: confirmMock }),
}))

const promptMock = vi.fn()
vi.mock('./usePrompt', () => ({
  usePrompt: () => ({ prompt: promptMock }),
}))

import { api } from '../api'
import {
  useSkillFileManager,
  fmtBytes,
  FILENAME_RE,
} from './useSkillFileManager'

const PLUGIN = 'demo'
const SKILL = 'my-skill'

function makeTextFile(name: string, content: string): File {
  const f = new File([content], name, { type: 'text/plain' })
  const encoder = new TextEncoder()
  Object.defineProperty(f, 'arrayBuffer', {
    value: () => Promise.resolve(encoder.encode(content).buffer),
  })
  return f
}

function makeBinaryFile(name: string, bytes: Uint8Array): File {
  const f = new File([bytes as BlobPart], name)
  Object.defineProperty(f, 'arrayBuffer', {
    value: () => Promise.resolve(bytes.buffer.slice(0)),
  })
  return f
}

function setup(opts: { onChanged?: () => Promise<void> | void } = {}) {
  const skillName = ref<string | null>(SKILL)
  const fm = useSkillFileManager(
    () => PLUGIN,
    () => skillName.value,
    opts,
  )
  return { fm, skillName }
}

beforeEach(() => {
  vi.clearAllMocks()
  confirmMock.mockResolvedValue(true)
  promptMock.mockResolvedValue(null)
})

describe('useSkillFileManager — helpers', () => {
  it('fmtBytes formats B/KB/MB', () => {
    expect(fmtBytes(0)).toBe('0 B')
    expect(fmtBytes(512)).toBe('512 B')
    expect(fmtBytes(2048)).toBe('2.0 KB')
    expect(fmtBytes(5 * 1024 * 1024)).toBe('5.00 MB')
  })

  it('FILENAME_RE accepts safe relative paths and rejects unsafe names', () => {
    expect(FILENAME_RE.test('build.py')).toBe(true)
    expect(FILENAME_RE.test('sub/util.sh')).toBe(true)
    expect(FILENAME_RE.test('a b.py')).toBe(false)
    expect(FILENAME_RE.test('a*b.py')).toBe(false)
    expect(FILENAME_RE.test('')).toBe(false)
  })
})

describe('useSkillFileManager — loadFiles', () => {
  it('populates files for the current skill', async () => {
    vi.mocked(api.listSkillFiles).mockResolvedValue([
      { path: 'scripts/a.py', sizeBytes: 10, isBinary: false, updatedAt: '' },
    ])
    const { fm } = setup()
    await fm.loadFiles()
    expect(fm.files.value).toHaveLength(1)
    expect(api.listSkillFiles).toHaveBeenCalledWith(PLUGIN, SKILL)
  })

  it('is a no-op when skillName is null (create mode)', async () => {
    const skillName = ref<string | null>(null)
    const fm = useSkillFileManager(() => PLUGIN, () => skillName.value)
    await fm.loadFiles()
    expect(api.listSkillFiles).not.toHaveBeenCalled()
  })

  it('sets fileError on rejection', async () => {
    vi.mocked(api.listSkillFiles).mockRejectedValue(new Error('boom'))
    const { fm } = setup()
    await fm.loadFiles()
    expect(fm.fileError.value).toBe('boom')
  })
})

describe('useSkillFileManager — selectFile', () => {
  it('loads file content and metadata', async () => {
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/a.py', content: 'print(1)', isBinary: false, sizeBytes: 8, updatedAt: '',
    })
    const { fm } = setup()
    await fm.selectFile('scripts/a.py')
    expect(fm.selectedPath.value).toBe('scripts/a.py')
    expect(fm.fileContent.value).toBe('print(1)')
    expect(fm.fileIsBinary.value).toBe(false)
    expect(fm.fileSize.value).toBe(8)
    expect(fm.fileLoading.value).toBe(false)
    expect(fm.fileDirty.value).toBe(false)
  })

  it('is a no-op when the same path is already selected', async () => {
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/a.py', content: 'x', isBinary: false, sizeBytes: 1, updatedAt: '',
    })
    const { fm } = setup()
    await fm.selectFile('scripts/a.py')
    vi.mocked(api.getSkillFile).mockClear()
    await fm.selectFile('scripts/a.py')
    expect(api.getSkillFile).not.toHaveBeenCalled()
  })
})

describe('useSkillFileManager — saveCurrentFile', () => {
  it('saves current content and triggers onChanged', async () => {
    const onChanged = vi.fn()
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/a.py', content: 'a', isBinary: false, sizeBytes: 1, updatedAt: '',
    })
    vi.mocked(api.putSkillFile).mockResolvedValue({
      path: 'scripts/a.py', content: '', isBinary: false, sizeBytes: 5, updatedAt: '',
    })
    vi.mocked(api.listSkillFiles).mockResolvedValue([])
    const { fm } = setup({ onChanged })
    await fm.selectFile('scripts/a.py')
    fm.fileContent.value = 'a = 1'
    fm.fileDirty.value = true
    await fm.saveCurrentFile()
    expect(api.putSkillFile).toHaveBeenCalledWith(PLUGIN, SKILL, 'scripts/a.py', {
      content: 'a = 1', isBinary: false,
    })
    expect(fm.fileDirty.value).toBe(false)
    expect(fm.fileSize.value).toBe(5)
    expect(onChanged).toHaveBeenCalled()
  })

  it('does nothing without a selected path', async () => {
    const { fm } = setup()
    await fm.saveCurrentFile()
    expect(api.putSkillFile).not.toHaveBeenCalled()
  })
})

describe('useSkillFileManager — deleteCurrentFile', () => {
  it('confirms, clears selection, and reloads', async () => {
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/a.py', content: 'x', isBinary: false, sizeBytes: 1, updatedAt: '',
    })
    vi.mocked(api.deleteSkillFile).mockResolvedValue(undefined)
    vi.mocked(api.listSkillFiles).mockResolvedValue([])
    const { fm } = setup()
    await fm.selectFile('scripts/a.py')
    await fm.deleteCurrentFile()
    expect(confirmMock).toHaveBeenCalled()
    expect(api.deleteSkillFile).toHaveBeenCalledWith(PLUGIN, SKILL, 'scripts/a.py')
    expect(fm.selectedPath.value).toBeNull()
  })

  it('is cancelled when confirm resolves false', async () => {
    confirmMock.mockResolvedValueOnce(false)
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/a.py', content: 'x', isBinary: false, sizeBytes: 1, updatedAt: '',
    })
    const { fm } = setup()
    await fm.selectFile('scripts/a.py')
    await fm.deleteCurrentFile()
    expect(api.deleteSkillFile).not.toHaveBeenCalled()
    expect(fm.selectedPath.value).toBe('scripts/a.py')
  })
})

describe('useSkillFileManager — promptNewFile', () => {
  it('rejects invalid filenames without API calls', async () => {
    promptMock.mockResolvedValueOnce('../bad name')
    const { fm } = setup()
    await fm.promptNewFile('scripts')
    expect(fm.fileError.value).toContain('invalid filename')
    expect(api.putSkillFile).not.toHaveBeenCalled()
  })

  it('is a no-op when the prompt is cancelled', async () => {
    promptMock.mockResolvedValueOnce(null)
    const { fm } = setup()
    await fm.promptNewFile('scripts')
    expect(api.putSkillFile).not.toHaveBeenCalled()
    expect(fm.fileError.value).toBe('')
  })

  it('is a no-op when the prompt is empty', async () => {
    promptMock.mockResolvedValueOnce('   ')
    const { fm } = setup()
    await fm.promptNewFile('scripts')
    expect(api.putSkillFile).not.toHaveBeenCalled()
    expect(fm.fileError.value).toBe('')
  })

  it('selects the existing file when path already present', async () => {
    promptMock.mockResolvedValueOnce('a.py')
    vi.mocked(api.listSkillFiles).mockResolvedValue([
      { path: 'scripts/a.py', sizeBytes: 1, isBinary: false, updatedAt: '' },
    ])
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/a.py', content: 'x', isBinary: false, sizeBytes: 1, updatedAt: '',
    })
    const { fm } = setup()
    await fm.loadFiles()
    await fm.promptNewFile('scripts')
    expect(api.putSkillFile).not.toHaveBeenCalled()
    expect(fm.selectedPath.value).toBe('scripts/a.py')
  })

  it('creates and selects a new file', async () => {
    promptMock.mockResolvedValueOnce('new.py')
    vi.mocked(api.putSkillFile).mockResolvedValue({
      path: 'scripts/new.py', content: '', isBinary: false, sizeBytes: 0, updatedAt: '',
    })
    vi.mocked(api.listSkillFiles).mockResolvedValue([])
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/new.py', content: '', isBinary: false, sizeBytes: 0, updatedAt: '',
    })
    const { fm } = setup()
    await fm.promptNewFile('scripts')
    expect(api.putSkillFile).toHaveBeenCalledWith(PLUGIN, SKILL, 'scripts/new.py', {
      content: '', isBinary: false,
    })
    expect(fm.selectedPath.value).toBe('scripts/new.py')
  })
})

describe('useSkillFileManager — uploadList', () => {
  it('skips invalid filenames and records an error', async () => {
    vi.mocked(api.listSkillFiles).mockResolvedValue([])
    const { fm } = setup()
    const bad = makeTextFile('has space.py', 'x')
    const input = { files: [bad] as unknown as FileList, value: 'x' } as HTMLInputElement
    await fm.onUploadChange('scripts', { target: input } as unknown as Event)
    expect(fm.fileError.value).toContain('skipped invalid filename')
    expect(api.putSkillFile).not.toHaveBeenCalled()
    expect(input.value).toBe('')
  })

  it('uploads valid text files and selects the last one', async () => {
    vi.mocked(api.putSkillFile).mockResolvedValue({
      path: 'scripts/ok.py', content: '', isBinary: false, sizeBytes: 0, updatedAt: '',
    })
    vi.mocked(api.listSkillFiles).mockResolvedValue([])
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/ok.py', content: 'hi', isBinary: false, sizeBytes: 2, updatedAt: '',
    })
    const { fm } = setup()
    const good = makeTextFile('ok.py', 'hi')
    const input = { files: [good] as unknown as FileList, value: 'x' } as HTMLInputElement
    await fm.onUploadChange('scripts', { target: input } as unknown as Event)
    expect(api.putSkillFile).toHaveBeenCalledWith(PLUGIN, SKILL, 'scripts/ok.py', {
      content: 'hi', isBinary: false,
    })
    expect(fm.selectedPath.value).toBe('scripts/ok.py')
    expect(input.value).toBe('')
  })

  it('uploads binary files as base64', async () => {
    vi.mocked(api.putSkillFile).mockResolvedValue({
      path: 'assets/img.bin', content: '', isBinary: true, sizeBytes: 3, updatedAt: '',
    })
    vi.mocked(api.listSkillFiles).mockResolvedValue([])
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'assets/img.bin', content: '', isBinary: true, sizeBytes: 3, updatedAt: '',
    })
    const { fm } = setup()
    // 0xFF 0xFE 0xFD is invalid UTF-8.
    const file = makeBinaryFile('img.bin', new Uint8Array([0xff, 0xfe, 0xfd]))
    const input = { files: [file] as unknown as FileList, value: '' } as HTMLInputElement
    await fm.onUploadChange('assets', { target: input } as unknown as Event)
    expect(api.putSkillFile).toHaveBeenCalledTimes(1)
    const call = vi.mocked(api.putSkillFile).mock.calls[0]
    expect(call[3].isBinary).toBe(true)
    expect(call[3].content).toBe(btoa(String.fromCharCode(0xff, 0xfe, 0xfd)))
  })
})

describe('useSkillFileManager — refreshAfterRevert', () => {
  it('keeps the selected path when it still exists after revert', async () => {
    vi.mocked(api.getSkillFile).mockResolvedValue({
      path: 'scripts/a.py', content: 'x', isBinary: false, sizeBytes: 1, updatedAt: '',
    })
    vi.mocked(api.listSkillFiles).mockResolvedValue([
      { path: 'scripts/a.py', sizeBytes: 1, isBinary: false, updatedAt: '' },
    ])
    const { fm } = setup()
    await fm.selectFile('scripts/a.py')
    vi.mocked(api.getSkillFile).mockClear()
    await fm.refreshAfterRevert()
    // selected path still valid → reloaded
    expect(api.getSkillFile).toHaveBeenCalledWith(PLUGIN, SKILL, 'scripts/a.py')
  })

  it('drops selection when the file is gone after revert', async () => {
    vi.mocked(api.getSkillFile).mockResolvedValueOnce({
      path: 'scripts/a.py', content: 'x', isBinary: false, sizeBytes: 1, updatedAt: '',
    })
    vi.mocked(api.listSkillFiles).mockResolvedValue([])
    const { fm } = setup()
    await fm.selectFile('scripts/a.py')
    await fm.refreshAfterRevert()
    expect(fm.selectedPath.value).toBeNull()
  })
})

describe('useSkillFileManager — triggerUpload', () => {
  it('clicks the matching input ref', async () => {
    const { fm } = setup()
    const click = vi.fn()
    fm.scriptsInput.value = { click } as unknown as HTMLInputElement
    fm.triggerUpload('scripts')
    expect(click).toHaveBeenCalled()
  })

  it('does nothing when the input ref is null', async () => {
    const { fm } = setup()
    fm.triggerUpload('scripts')
    await nextTick()
    // no throw
  })
})
