package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type fileEntry struct {
	status string
	path   string
}

func (f fileEntry) Title() string       { return f.path }
func (f fileEntry) Description() string { return "" }
func (f fileEntry) FilterValue() string { return f.path }

func (f fileEntry) StatusLabel() string {
	return statusLabel(f.status)
}

func getCurrentBranch() string {
	out, err := gitCmd("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(out)
}

func gitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func getChangedFiles() ([]fileEntry, error) {
	// Use -uall to expand untracked directories into individual files
	out, err := gitCmd("status", "--porcelain", "-uall")
	if err != nil {
		return nil, err
	}
	var files []fileEntry
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if len(line) < 4 {
			continue
		}
		status := strings.TrimSpace(line[:2])
		path := strings.TrimSpace(line[3:])
		// Handle renames: "R  old -> new"
		if idx := strings.Index(path, " -> "); idx != -1 {
			path = path[idx+4:]
		}
		files = append(files, fileEntry{status: status, path: path})
	}
	return files, nil
}

func getAllFiles() ([]fileEntry, error) {
	out, err := gitCmd("ls-files")
	if err != nil {
		return nil, err
	}
	var files []fileEntry
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		files = append(files, fileEntry{status: " ", path: line})
	}
	return files, nil
}

func getDiff(path string) (string, error) {
	// Try staged diff first, then unstaged
	out, err := gitCmd("diff", "--cached", "--", path)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) != "" {
		return out, nil
	}
	out, err = gitCmd("diff", "--", path)
	if err != nil {
		return "", err
	}
	return out, nil
}

func readFile(path string) (string, error) {
	full := filepath.Join(workDir, path)
	info, err := os.Stat(full)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return listDir(full)
	}
	out, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func listDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		b.WriteString(name + "\n")
	}
	return b.String(), nil
}

func isBinary(data string) bool {
	// Check first 8KB for null bytes or invalid UTF-8
	sample := data
	if len(sample) > 8192 {
		sample = sample[:8192]
	}
	if strings.Contains(sample, "\x00") {
		return true
	}
	return !utf8.ValidString(sample)
}
