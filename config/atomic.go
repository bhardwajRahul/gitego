package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// atomicWriteFile publishes complete files only. The temporary file is placed
// beside the destination so the final replacement cannot cross filesystems.
func atomicWriteFile(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPermissions); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".gitego-tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()
	if err := f.Chmod(mode); err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := replaceFile(tmp, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return syncDirectory(dir)
}
