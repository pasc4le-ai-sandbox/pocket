package pocket

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTest creates a temporary home directory and returns a cleanup function.
// All operations inside the test will use this temp dir as the user's home,
// so ~/.pocketrc is scoped to the test.
func setupTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Also clear XDG on macOS/Linux SSH scenarios
	t.Setenv("USERPROFILE", dir)
	return dir
}

// writeTestFile creates a file with content inside dir.
func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// ---------------------------------------------------------------------------
// dataPath
// ---------------------------------------------------------------------------

func TestDataPath(t *testing.T) {
	home := setupTest(t)
	p, err := dataPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".pocketrc")
	if p != want {
		t.Errorf("dataPath() = %q, want %q", p, want)
	}
}

// ---------------------------------------------------------------------------
// save / load
// ---------------------------------------------------------------------------

func TestSaveAndLoad(t *testing.T) {
	setupTest(t)

	// Empty save/load
	if err := save([]string{}); err != nil {
		t.Fatal(err)
	}
	got, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("load after empty save: got %d items, want 0", len(got))
	}

	// Non-empty save/load
	items := []string{"/a/b", "/c/d"}
	if err := save(items); err != nil {
		t.Fatal(err)
	}
	got, err = load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "/a/b" || got[1] != "/c/d" {
		t.Errorf("load after save = %v, want %v", got, items)
	}
}

func TestLoadMissingFile(t *testing.T) {
	setupTest(t)

	items, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("load on missing file = %v, want empty", items)
	}
}

// ---------------------------------------------------------------------------
// resolvePaths
// ---------------------------------------------------------------------------

func TestResolvePaths(t *testing.T) {
	home := setupTest(t)

	f1 := writeTestFile(t, home, "a.txt", "hello")
	f2 := writeTestFile(t, home, "b.txt", "world")
	_ = os.MkdirAll(filepath.Join(home, "sub"), 0755)

	tests := []struct {
		name    string
		args    []string
		want    []string
		wantErr bool
	}{
		{"single", []string{f1}, []string{f1}, false},
		{"multiple", []string{f1, f2}, []string{f1, f2}, false},
		{"nonexistent", []string{"/no/such/file"}, nil, true},
		{"directory", []string{filepath.Join(home, "sub")}, []string{filepath.Join(home, "sub")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePaths(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePaths() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("resolvePaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Add / List / Delete integration
// ---------------------------------------------------------------------------

func TestAddList(t *testing.T) {
	home := setupTest(t)
	f1 := writeTestFile(t, home, "a.txt", "hello")
	f2 := writeTestFile(t, home, "b.txt", "world")

	// Add two files
	if err := Add(f1); err != nil {
		t.Fatal(err)
	}
	if err := Add(f2); err != nil {
		t.Fatal(err)
	}

	items, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("List() returned %d items, want 2", len(items))
	}

	// Both should be absolute paths
	if !filepath.IsAbs(items[0]) {
		t.Errorf("item[0] is not absolute: %s", items[0])
	}
}

func TestAddNonExistent(t *testing.T) {
	setupTest(t)
	if err := Add("/no/such/file.txt"); err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestAddMultiple(t *testing.T) {
	home := setupTest(t)
	f1 := writeTestFile(t, home, "a.txt", "hello")
	f2 := writeTestFile(t, home, "b.txt", "world")
	f3 := writeTestFile(t, home, "c.txt", "!")

	if err := Add(f1, f2, f3); err != nil {
		t.Fatal(err)
	}
	items, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("List() = %d items, want 3", len(items))
	}
}

func TestDelete(t *testing.T) {
	home := setupTest(t)
	f1 := writeTestFile(t, home, "a.txt", "hello")
	f2 := writeTestFile(t, home, "b.txt", "world")
	f3 := writeTestFile(t, home, "c.txt", "!")

	if err := Add(f1, f2, f3); err != nil {
		t.Fatal(err)
	}

	// Delete middle item (2)
	if err := Delete(2); err != nil {
		t.Fatal(err)
	}
	items, _ := List()
	if len(items) != 2 {
		t.Fatalf("after delete: %d items, want 2", len(items))
	}
	// First and last should remain
	if filepath.Base(items[0]) != "a.txt" || filepath.Base(items[1]) != "c.txt" {
		t.Errorf("after delete item 2: got %v, want [a.txt, c.txt]", items)
	}

	// Delete out of range
	if err := Delete(5); err == nil {
		t.Fatal("expected error for out-of-range delete")
	}
	if err := Delete(0); err == nil {
		t.Fatal("expected error for delete(0)")
	}
	if err := Delete(-1); err == nil {
		t.Fatal("expected error for delete(-1)")
	}

	// Delete first
	if err := Delete(1); err != nil {
		t.Fatal(err)
	}
	items, _ = List()
	if len(items) != 1 {
		t.Fatalf("after delete first: %d items, want 1", len(items))
	}

	// Delete last (only remaining)
	if err := Delete(1); err != nil {
		t.Fatal(err)
	}
	items, _ = List()
	if len(items) != 0 {
		t.Errorf("after deleting all: %d items, want 0", len(items))
	}
}

// ---------------------------------------------------------------------------
// Release
// ---------------------------------------------------------------------------

func TestReleaseEmpty(t *testing.T) {
	setupTest(t)

	err := Release(false)
	if err == nil {
		t.Fatal("expected error for empty clipboard release")
	}
}

func TestReleaseCopy(t *testing.T) {
	home := setupTest(t)
	f1 := writeTestFile(t, home, "a.txt", "hello")
	f2 := writeTestFile(t, home, "b.txt", "world")
	sub := filepath.Join(home, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, sub, "inner.txt", "nested")

	if err := Add(f1, f2, sub); err != nil {
		t.Fatal(err)
	}

	// Create a release destination and cd into it
	dst := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dst); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	if err := Release(false); err != nil {
		t.Fatal(err)
	}

	// Verify files were copied
	for _, name := range []string{"a.txt", "b.txt", "sub"} {
		if _, err := os.Stat(filepath.Join(dst, name)); err != nil {
			t.Errorf("released file %s not found: %v", name, err)
		}
	}

	// Verify a.txt content
	data, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(data) != "hello" {
		t.Errorf("a.txt content = %q, want %q", string(data), "hello")
	}

	// Verify nested content
	data, _ = os.ReadFile(filepath.Join(dst, "sub", "inner.txt"))
	if string(data) != "nested" {
		t.Errorf("sub/inner.txt content = %q, want %q", string(data), "nested")
	}

	// Verify clipboard was cleared
	items, _ := List()
	if len(items) != 0 {
		t.Errorf("clipboard not cleared after release: %d items", len(items))
	}
}

func TestReleaseCut(t *testing.T) {
	home := setupTest(t)
	f1 := writeTestFile(t, home, "moveme.txt", "bye")
	sub := filepath.Join(home, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, sub, "nested.txt", "inside")

	if err := Add(f1, sub); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dst); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	if err := Release(true); err != nil {
		t.Fatal(err)
	}

	// Verify files exist at destination
	for _, name := range []string{"moveme.txt", "sub"} {
		if _, err := os.Stat(filepath.Join(dst, name)); err != nil {
			t.Errorf("moved %s not found: %v", name, err)
		}
	}

	// Verify originals are gone
	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Errorf("original %s should be gone after cut", f1)
	}
	if _, err := os.Stat(sub); !os.IsNotExist(err) {
		t.Errorf("original %s should be gone after cut", sub)
	}
}

func TestReleaseDestinationConflict(t *testing.T) {
	home := setupTest(t)
	f1 := writeTestFile(t, home, "a.txt", "original")
	if err := Add(f1); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	// Pre-create a.txt in the destination directory
	writeTestFile(t, dst, "a.txt", "preexisting")

	orig, _ := os.Getwd()
	if err := os.Chdir(dst); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	err := Release(false)
	if err == nil {
		t.Fatal("expected error due to destination conflict")
	}

	// The pre-existing file should remain untouched
	data, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(data) != "preexisting" {
		t.Errorf("conflicting file was overwritten: content = %q, want %q", string(data), "preexisting")
	}
}

// ---------------------------------------------------------------------------
// copyFile / copyDir
// ---------------------------------------------------------------------------

func TestCopyFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.txt")
	if err := os.WriteFile(src, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "dst.txt")

	info, _ := os.Stat(src)
	if err := copyFile(src, dst, info); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(dst)
	if string(data) != "hello" {
		t.Errorf("copyFile content = %q, want %q", string(data), "hello")
	}
}

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	writeTestFile(t, src, "a.txt", "A")
	writeTestFile(t, src, "b.txt", "B")
	sub := filepath.Join(src, "sub")
	os.MkdirAll(sub, 0755)
	writeTestFile(t, sub, "c.txt", "C")

	dst := filepath.Join(t.TempDir(), "copydir")
	if err := copyDir(src, dst); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{"a.txt", "b.txt", "sub/c.txt"} {
		full := filepath.Join(dst, p)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("copied path %s not found: %v", p, err)
		}
	}
}

func TestCopyOrMove(t *testing.T) {
	dir := t.TempDir()
	src := writeTestFile(t, dir, "file.txt", "data")
	dst := filepath.Join(dir, "moved.txt")

	// Copy
	if err := copyOrMove(src, dst, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Errorf("source should still exist after copy: %v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("destination should exist after copy: %v", err)
	}

	// Move
	dst2 := filepath.Join(dir, "moved2.txt")
	if err := copyOrMove(src, dst2, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should be gone after move: %v", err)
	}
	if _, err := os.Stat(dst2); err != nil {
		t.Errorf("destination should exist after move: %v", err)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
