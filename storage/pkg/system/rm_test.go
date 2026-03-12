package system

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.podman.io/storage/pkg/mount"
)

func TestEnsureRemoveAllNotExist(t *testing.T) {
	// should never return an error for a non-existent path
	if err := EnsureRemoveAll("/non/existent/path"); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureRemoveAllWithDir(t *testing.T) {
	dir := t.TempDir()
	if err := EnsureRemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureRemoveAllWithFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test-ensure-removeall-with-file")
	err := os.WriteFile(tmp, []byte{}, 0o644)
	require.NoError(t, err)
	if err := EnsureRemoveAll(tmp); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureRemoveAllWithMount(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("mount not supported on Windows")
	}

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	bindDir := filepath.Join(dir1, "bind")
	if err := os.MkdirAll(bindDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := mount.Mount(dir2, bindDir, "none", "bind"); err != nil {
		t.Fatal(err)
	}

	var err error
	done := make(chan struct{})
	go func() {
		err = EnsureRemoveAll(dir1)
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for EnsureRemoveAll to finish")
	}

	if _, err := os.Stat(dir1); !os.IsNotExist(err) {
		t.Fatalf("expected %q to not exist", dir1)
	}
}
