package formatgo

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/rinchsan/gosimports"
	"golang.org/x/mod/modfile"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/gitstatus"
)

var rxCodeGenerated = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func Run() error {
	return FormatFiles(getFormatFiles())
}

func FormatFiles(paths []string) error {
	return formatFiles(paths, config.C.Format.DisplayDiff, os.Stdout)
}

func formatFiles(paths []string, displayDiff bool, output io.Writer) error {
	files, err := collectGoFiles(paths)
	if err != nil {
		return err
	}

	var changed int
	for _, file := range files {
		original, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if isGenerated(original) {
			continue
		}

		formatted, err := formatSource(file, original)
		if err != nil {
			return fmt.Errorf("format %s: %w", file, err)
		}
		if bytes.Equal(original, formatted) {
			continue
		}

		changed++
		if displayDiff {
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(original)),
				B:        difflib.SplitLines(string(formatted)),
				FromFile: file,
				ToFile:   file,
				Context:  3,
			}
			text, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				return err
			}
			if _, err := io.WriteString(output, text); err != nil {
				return err
			}
			continue
		}

		info, err := os.Stat(file)
		if err != nil {
			return err
		}
		if err := os.WriteFile(file, formatted, info.Mode().Perm()); err != nil {
			return err
		}
	}

	if displayDiff && changed > 0 {
		return fmt.Errorf("%d files should be formatted", changed)
	}
	return nil
}

func formatSource(filename string, source []byte) ([]byte, error) {
	cmd := exec.Command("gofmt", "-s")
	cmd.Stdin = bytes.NewReader(source)
	formatted, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gofmt: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}

	localPrefix, err := findModulePath(filename)
	if err != nil {
		return nil, err
	}
	gosimports.LocalPrefix = localPrefix
	return gosimports.Process(filename, formatted, &gosimports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: true,
	})
}

func collectGoFiles(paths []string) ([]string, error) {
	seen := make(map[string]struct{})
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if filepath.Ext(path) == ".go" {
				seen[filepath.Clean(path)] = struct{}{}
			}
			continue
		}

		err = filepath.WalkDir(path, func(filename string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() && filename != path {
				switch entry.Name() {
				case ".git", "vendor":
					return filepath.SkipDir
				}
			}
			if !entry.IsDir() && filepath.Ext(filename) == ".go" {
				seen[filepath.Clean(filename)] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	files := make([]string, 0, len(seen))
	for file := range seen {
		files = append(files, file)
	}
	sort.Strings(files)
	return files, nil
}

func findModulePath(filename string) (string, error) {
	dir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return "", err
	}
	for {
		goMod := filepath.Join(dir, "go.mod")
		data, readErr := os.ReadFile(goMod)
		if readErr == nil {
			parsed, err := modfile.Parse(goMod, data, nil)
			if err != nil {
				return "", err
			}
			if parsed.Module == nil {
				return "", fmt.Errorf("module directive not found in %s", goMod)
			}
			return parsed.Module.Mod.Path, nil
		}
		if !os.IsNotExist(readErr) {
			return "", readErr
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func isGenerated(source []byte) bool {
	line, _, _ := bytes.Cut(source, []byte("\n"))
	return rxCodeGenerated.Match(line)
}

func getFormatFiles() []string {
	if config.C.Format.GitChange {
		files, _, err := gitstatus.ChangedFiles(".", ".go")
		if err == nil {
			return files
		}
	}
	return []string{"."}
}
