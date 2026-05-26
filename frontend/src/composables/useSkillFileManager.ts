import { ref, computed, toValue, type MaybeRefOrGetter } from 'vue'
import { api, errMsg } from '../api'
import type { SkillFileSummary } from '../types'
import { useConfirm } from './useConfirm'
import { usePrompt } from './usePrompt'

// A folder is the first "/"-separated segment of a SkillFile path. The 3
// "well-known" folders below get default UI affordances (intro hint, + new
// button). Any other folder name accepted by the backend (via the API tool)
// is rendered too once files exist under it — see folderList computation.
export type SkillFolder = string

export const FOLDER_ORDER: readonly SkillFolder[] = ['scripts', 'references', 'assets']

export const FOLDER_HINT: Record<string, string> = {
  scripts: 'Code Claude can run (Python, bash, …)',
  references: 'Reference docs Claude reads on demand',
  assets: 'Templates, fonts, icons used in output',
}

export function isWellKnownFolder(folder: string): boolean {
  return folder in FOLDER_HINT
}

export const FILENAME_RE = /^[A-Za-z0-9_.-]+(\/[A-Za-z0-9_.-]+)*$/

// sanitizeUploadedFilename maps a user-chosen filename (e.g. from a file
// dialog) onto the restricted character set the backend accepts for skill file
// paths. Spaces and other punctuation become "_" so common drops like
// "Screenshot 2026-03-14 at 9.45.34 PM.png" upload cleanly. Returns the empty
// string if nothing usable is left (caller should reject).
export function sanitizeUploadedFilename(name: string): string {
  const stripped = name.replace(/^.*[\\/]/, '')
  const replaced = stripped.replace(/[^A-Za-z0-9_.-]+/g, '_')
  if (/^\.+$/.test(replaced)) return ''
  return replaced
}

export function fmtBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / (1024 * 1024)).toFixed(2)} MB`
}

function isProbablyUtf8(bytes: Uint8Array): boolean {
  try {
    new TextDecoder('utf-8', { fatal: true }).decode(bytes)
    return true
  } catch {
    return false
  }
}

function base64FromBytes(bytes: Uint8Array): string {
  // Chunk to avoid call stack limits on large files.
  let binary = ''
  const chunk = 0x8000
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk))
  }
  return btoa(binary)
}

export interface UseSkillFileManagerOptions {
  onChanged?: () => Promise<void> | void
}

export function useSkillFileManager(
  pluginName: MaybeRefOrGetter<string>,
  skillName: MaybeRefOrGetter<string | null>,
  options: UseSkillFileManagerOptions = {},
) {
  const { confirm } = useConfirm()
  const { prompt } = usePrompt()
  const onChanged = options.onChanged

  const files = ref<SkillFileSummary[]>([])
  const selectedPath = ref<string | null>(null)
  const fileContent = ref('')
  const fileIsBinary = ref(false)
  const fileSize = ref(0)
  const fileLoading = ref(false)
  const fileDirty = ref(false)
  const fileError = ref('')

  // Single shared file input — when the user clicks "upload" on a folder
  // header, triggerUpload sets pendingUploadFolder and forwards the click.
  // Each folder used to own its own input ref; consolidating to one keeps the
  // template simple now that folders are dynamic.
  const uploadInput = ref<HTMLInputElement | null>(null)
  const pendingUploadFolder = ref<string | null>(null)

  const filesByFolder = computed(() => {
    const out: Record<string, SkillFileSummary[]> = {}
    for (const folder of FOLDER_ORDER) out[folder] = []
    for (const f of files.value) {
      const root = f.path.split('/', 1)[0]
      if (!out[root]) out[root] = []
      out[root].push(f)
    }
    return out
  })

  const folderList = computed<string[]>(() => {
    const known = new Set<string>(FOLDER_ORDER)
    const extras: string[] = []
    for (const folder of Object.keys(filesByFolder.value)) {
      if (!known.has(folder)) extras.push(folder)
    }
    extras.sort()
    return [...FOLDER_ORDER, ...extras]
  })

  function requireSkill(): { plugin: string; skill: string } | null {
    const skill = toValue(skillName)
    if (!skill) return null
    return { plugin: toValue(pluginName), skill }
  }

  async function afterMutation() {
    await loadFiles()
    if (onChanged) await onChanged()
  }

  async function loadFiles() {
    const ctx = requireSkill()
    if (!ctx) return
    try {
      files.value = await api.listSkillFiles(ctx.plugin, ctx.skill)
    } catch (e: unknown) {
      fileError.value = errMsg(e)
    }
  }

  function clearFileSelection() {
    selectedPath.value = null
    fileError.value = ''
    fileDirty.value = false
  }

  async function loadSelectedFile(path: string) {
    const ctx = requireSkill()
    if (!ctx) return
    selectedPath.value = path
    fileError.value = ''
    fileDirty.value = false
    fileLoading.value = true
    try {
      const f = await api.getSkillFile(ctx.plugin, ctx.skill, path)
      fileContent.value = f.content
      fileIsBinary.value = f.isBinary
      fileSize.value = f.sizeBytes
    } catch (e: unknown) {
      fileError.value = errMsg(e)
    } finally {
      fileLoading.value = false
    }
  }

  async function selectFile(path: string) {
    if (selectedPath.value === path) return
    await loadSelectedFile(path)
  }

  async function saveCurrentFile() {
    const ctx = requireSkill()
    if (!ctx || !selectedPath.value) return
    fileError.value = ''
    try {
      const saved = await api.putSkillFile(
        ctx.plugin,
        ctx.skill,
        selectedPath.value,
        { content: fileContent.value, isBinary: fileIsBinary.value },
      )
      fileSize.value = saved.sizeBytes
      fileDirty.value = false
      await afterMutation()
    } catch (e: unknown) {
      fileError.value = errMsg(e)
    }
  }

  async function deleteCurrentFile() {
    const ctx = requireSkill()
    if (!ctx || !selectedPath.value) return
    const ok = await confirm({
      title: 'Delete file',
      message: `Delete ${selectedPath.value}? This creates a new version, which you can revert if needed.`,
      confirmLabel: 'Delete',
      danger: true,
    })
    if (!ok) return
    try {
      await api.deleteSkillFile(ctx.plugin, ctx.skill, selectedPath.value)
      clearFileSelection()
      await afterMutation()
    } catch (e: unknown) {
      fileError.value = errMsg(e)
    }
  }

  function downloadCurrentFile() {
    if (!selectedPath.value) return
    let blob: Blob
    if (fileIsBinary.value) {
      const bin = atob(fileContent.value)
      const bytes = new Uint8Array(bin.length)
      for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i)
      blob = new Blob([bytes], { type: 'application/octet-stream' })
    } else {
      blob = new Blob([fileContent.value], { type: 'text/plain;charset=utf-8' })
    }
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = selectedPath.value.split('/').pop() || 'file'
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  }

  async function promptNewFile(folder: SkillFolder) {
    const ctx = requireSkill()
    if (!ctx) return
    const raw = await prompt({
      title: `New file in ${folder}/`,
      message: 'Enter a relative path (e.g. build.py or sub/util.sh).',
      placeholder: 'build.py',
      confirmLabel: 'Create',
    })
    if (raw === null) return
    const trimmed = raw.trim().replace(/^\/+/, '')
    if (!trimmed) return
    if (!FILENAME_RE.test(trimmed)) {
      fileError.value = `invalid filename: ${trimmed}`
      return
    }
    const path = `${folder}/${trimmed}`
    if (files.value.some(f => f.path === path)) {
      await selectFile(path)
      return
    }
    try {
      await api.putSkillFile(ctx.plugin, ctx.skill, path, {
        content: '',
        isBinary: false,
      })
      await afterMutation()
      await selectFile(path)
    } catch (e: unknown) {
      fileError.value = errMsg(e)
    }
  }

  function triggerUpload(folder: SkillFolder) {
    pendingUploadFolder.value = folder
    uploadInput.value?.click()
  }

  async function uploadList(folder: SkillFolder, list: FileList) {
    const ctx = requireSkill()
    if (!ctx) return
    fileError.value = ''
    let lastPath: string | null = null
    for (const file of Array.from(list)) {
      const safe = sanitizeUploadedFilename(file.name)
      if (!safe || !FILENAME_RE.test(safe)) {
        fileError.value = `skipped invalid filename: ${file.name}`
        continue
      }
      const path = `${folder}/${safe}`
      try {
        const buf = await file.arrayBuffer()
        const bytes = new Uint8Array(buf)
        const binary = !isProbablyUtf8(bytes)
        const content = binary
          ? base64FromBytes(bytes)
          : new TextDecoder().decode(bytes)
        await api.putSkillFile(ctx.plugin, ctx.skill, path, {
          content,
          isBinary: binary,
        })
        lastPath = path
      } catch (e: unknown) {
        fileError.value = `${file.name}: ${errMsg(e)}`
      }
    }
    await afterMutation()
    if (lastPath) await selectFile(lastPath)
  }

  async function onUploadChange(ev: Event) {
    const input = ev.target as HTMLInputElement
    const folder = pendingUploadFolder.value
    pendingUploadFolder.value = null
    if (!input.files || !folder) return
    await uploadList(folder, input.files)
    input.value = ''
  }

  async function onDrop(folder: SkillFolder, ev: DragEvent) {
    ev.preventDefault()
    if (!ev.dataTransfer?.files) return
    await uploadList(folder, ev.dataTransfer.files)
  }

  async function refreshAfterRevert() {
    await loadFiles()
    if (selectedPath.value) {
      if (files.value.some(f => f.path === selectedPath.value)) {
        await loadSelectedFile(selectedPath.value)
      } else {
        clearFileSelection()
      }
    }
  }

  return {
    files,
    selectedPath,
    fileContent,
    fileIsBinary,
    fileSize,
    fileLoading,
    fileDirty,
    fileError,
    filesByFolder,
    folderList,
    uploadInput,
    loadFiles,
    clearFileSelection,
    selectFile,
    saveCurrentFile,
    deleteCurrentFile,
    downloadCurrentFile,
    promptNewFile,
    triggerUpload,
    onUploadChange,
    onDrop,
    refreshAfterRevert,
  }
}
