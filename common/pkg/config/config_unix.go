//go:build !windows

package config

import (
	"path/filepath"
	"golang.org/x/sys/unix"
)

var ErrDiskFull = unix.ENOSPC

func safeEvalSymlinks(filePath string) (string, error) {
	return filepath.EvalSymlinks(filePath)
}
