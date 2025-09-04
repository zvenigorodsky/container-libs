//go:build !windows && !linux

package timezone

import (
	"golang.org/x/sys/unix"
)

func openDirectory(path string) (fd int, err error) {
	// FIXME: If O_PATH is not defined on a platform, it probably doesn't work. E.g. on macOS, this is actually O_DSYNC.
	const O_PATH = 0x00400000 //nolint:staticcheck // ST1003: should not use ALL_CAPS
	return unix.Open(path, unix.O_RDONLY|O_PATH|unix.O_CLOEXEC, 0)
}
