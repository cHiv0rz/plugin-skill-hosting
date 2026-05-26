// Command import-plugin uploads an on-disk plugin directory into a running
// marketplace instance through its REST API. Point it at a directory that
// contains .claude-plugin/plugin.json (the format the server materialises into
// git) and it will create the plugin and import every skill under skills/.
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type pluginManifest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Author      *manifestAuthor `json:"author"`
	Homepage    string          `json:"homepage"`
	License     string          `json:"license"`
}

type manifestAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type createPluginReq struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	AuthorName  string `json:"authorName,omitempty"`
	AuthorEmail string `json:"authorEmail,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	License     string `json:"license,omitempty"`
}

func main() {
	baseURL := flag.String("url", os.Getenv("MARKETPLACE_URL"), "marketplace base URL (env: MARKETPLACE_URL)")
	token := flag.String("token", os.Getenv("MARKETPLACE_TOKEN"), "API bearer token (env: MARKETPLACE_TOKEN)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [--url URL] [--token TOKEN] <plugin-dir>\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(os.Stderr, "Imports an on-disk plugin into a marketplace instance via the REST API.")
		fmt.Fprintln(os.Stderr, "The directory must contain .claude-plugin/plugin.json; any skills under")
		fmt.Fprintln(os.Stderr, "skills/<name>/ are zipped and imported one by one.")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if *baseURL == "" {
		fatal("missing --url (or MARKETPLACE_URL)")
	}
	if *token == "" {
		fatal("missing --token (or MARKETPLACE_TOKEN)")
	}
	if u, err := url.Parse(*baseURL); err != nil || u.Scheme == "" || u.Host == "" {
		fatal("invalid --url %q", *baseURL)
	}
	if err := run(strings.TrimRight(*baseURL, "/"), *token, flag.Arg(0)); err != nil {
		fatal("%v", err)
	}
}

func run(baseURL, token, pluginDir string) error {
	manifest, err := readManifest(pluginDir)
	if err != nil {
		return fmt.Errorf("read plugin manifest: %w", err)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return errors.New("plugin.json is missing 'name'")
	}

	req := createPluginReq{
		Name:        manifest.Name,
		Description: manifest.Description,
		Homepage:    manifest.Homepage,
		License:     manifest.License,
	}
	if manifest.Author != nil {
		req.AuthorName = manifest.Author.Name
		req.AuthorEmail = manifest.Author.Email
	}

	fmt.Printf("Creating plugin %q...\n", manifest.Name)
	if err := createPlugin(baseURL, token, req); err != nil {
		return fmt.Errorf("create plugin: %w", err)
	}

	skillDirs, err := listSkillDirs(pluginDir)
	if err != nil {
		return fmt.Errorf("list skills: %w", err)
	}
	if len(skillDirs) == 0 {
		fmt.Println("No skills/ subdirectories found; nothing else to import.")
		return nil
	}
	for _, sd := range skillDirs {
		name := filepath.Base(sd)
		fmt.Printf("Importing skill %q...\n", name)
		buf, err := zipSkillDir(sd)
		if err != nil {
			return fmt.Errorf("zip skill %s: %w", name, err)
		}
		if err := importSkill(baseURL, token, manifest.Name, buf); err != nil {
			return fmt.Errorf("import skill %s: %w", name, err)
		}
	}
	fmt.Printf("Done. Imported %d skill(s).\n", len(skillDirs))
	return nil
}

func readManifest(pluginDir string) (*pluginManifest, error) {
	p := filepath.Join(pluginDir, ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var m pluginManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", p, err)
	}
	return &m, nil
}

func listSkillDirs(pluginDir string) ([]string, error) {
	skillsRoot := filepath.Join(pluginDir, "skills")
	entries, err := os.ReadDir(skillsRoot)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, filepath.Join(skillsRoot, e.Name()))
		}
	}
	return out, nil
}

func zipSkillDir(dir string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	walkErr := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		hdr.Method = zip.Deflate
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, f)
		f.Close()
		return copyErr
	})
	if walkErr != nil {
		return nil, walkErr
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

func createPlugin(baseURL, token string, req createPluginReq) error {
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/plugins", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	return doRequest(httpReq)
}

func importSkill(baseURL, token, pluginName string, zipBuf *bytes.Buffer) error {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "skill.zip")
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, zipBuf); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}
	httpReq, err := http.NewRequest(http.MethodPost,
		baseURL+"/api/plugins/"+url.PathEscape(pluginName)+"/skills/import", &body)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	return doRequest(httpReq)
}

func doRequest(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		var parsed struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(raw, &parsed) == nil && parsed.Error != "" {
			return fmt.Errorf("server returned %d: %s", resp.StatusCode, parsed.Error)
		}
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, bytes.TrimSpace(raw))
	}
	io.Copy(io.Discard, resp.Body)
	return nil
}

func fatal(format string, args ...any) {
	fmt.Fprintln(os.Stderr, "import-plugin: "+fmt.Sprintf(format, args...))
	os.Exit(1)
}
