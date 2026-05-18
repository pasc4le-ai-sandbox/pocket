// Package pocket implements a file-based clipboard: paths are saved to a
// JSON store (~/.pocketrc) and can be listed, selectively removed, or
// batch-released (copied/moved) to the current working directory.
package pocket

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const dataFileName = ".pocketrc"

// dataPath returns the path to the persistent store.
func dataPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	return filepath.Join(home, dataFileName), nil
}

// load reads the store from disk; returns an empty slice if missing.
func load() ([]string, error) {
	p, err := dataPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", p, err)
	}
	var items []string
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", p, err)
	}
	return items, nil
}

// save writes items to the store.
func save(items []string) error {
	p, err := dataPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding: %w", err)
	}
	if err := os.WriteFile(p, b, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", p, err)
	}
	return nil
}

// resolvePaths expands glob patterns and checks that every path exists.
func resolvePaths(raw []string) ([]string, error) {
	var resolved []string
	for _, r := range raw {
		matches, err := filepath.Glob(r)
		if err != nil {
			return nil, fmt.Errorf("bad pattern %q: %w", r, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("%s: no matching files", r)
		}
		// Sort so output is deterministic.
		slices.Sort(matches)
		resolved = append(resolved, matches...)
	}
	return resolved, nil
}

// Add appends one or more paths to the clipboard. Paths are stored as
// absolute paths so release works from any working directory.
func Add(raw ...string) error {
	paths, err := resolvePaths(raw)
	if err != nil {
		return err
	}
	abs := make([]string, len(paths))
	for i, p := range paths {
		a, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolving %s: %w", p, err)
		}
		abs[i] = a
	}
	items, err := load()
	if err != nil {
		return err
	}
	items = append(items, abs...)
	return save(items)
}

// List returns the current clipboard contents.
func List() ([]string, error) {
	return load()
}

// Delete removes the item at position num (1-indexed). Returns an error for
// out-of-range values. This only removes the entry from the clipboard — the
// underlying file is never touched.
func Delete(num int) error {
	items, err := load()
	if err != nil {
		return err
	}
	if num < 1 || num > len(items) {
		return fmt.Errorf("item %d not in clipboard (have %d items)", num, len(items))
	}
	items = append(items[:num-1], items[num:]...)
	return save(items)
}

// Release copies (or moves, when cut is true) every item in the clipboard to
// the current working directory. When keep is false the clipboard is cleared
// on full success; set keep to true to retain the items after release.
func Release(cut, keep bool) error {
	items, err := load()
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("clipboard is empty")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get working directory: %w", err)
	}

	var copied int
	var errors []string
	for _, src := range items {
		base := filepath.Base(src)
		dst := filepath.Join(cwd, base)

		// Check for name collisions.
		if _, err := os.Stat(dst); err == nil {
			errors = append(errors, fmt.Sprintf("%s -> %s: destination exists, skipping", src, dst))
			continue
		}

		if err := copyOrMove(src, dst, cut); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", src, err))
			continue
		}
		copied++
	}

	// Only clear the clipboard on full success unless --keep is set.
	if !keep && len(errors) == 0 {
		if err := save([]string{}); err != nil {
			return err
		}
	}

	if copied > 0 {
		verb := "Copied"
		if cut {
			verb = "Moved"
		}
		fmt.Printf("%s %d item(s) to %s\n", verb, copied, cwd)
	}
	if len(errors) > 0 {
		return fmt.Errorf("%d error(s):\n%s", len(errors), strings.Join(errors, "\n"))
	}
	return nil
}

// copyOrMove copies or moves a file/directory from src to dst.
func copyOrMove(src, dst string, move bool) error {
	if move {
		return os.Rename(src, dst)
	}
	return copyTree(src, dst)
}

// copyTree recursively copies a file or directory.
func copyTree(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst, info)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if err := copyTree(s, d); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, info os.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
