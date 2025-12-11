//go:build !windows

package dlhelper

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

// Acquire lock
func (fl *FileLocker) Acquire() error {
	f, err := os.Create(fl.path)
	if err != nil {
		return err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		f.Close()
		return fmt.Errorf("failed to lock file: %w", err)
	}
	fl.file = f
	return nil
}

// Release lock
func (fl *FileLocker) Release() {
	if err := unix.Flock(int(fl.file.Fd()), unix.LOCK_UN); err != nil {
		log.Err(err).Str("path", fl.path).Msg("failed to unlock file")
	}
	if err := fl.file.Close(); err != nil {
		log.Err(err).Str("path", fl.path).Msg("failed to close file")
	}
	if err := os.Remove(fl.path); err != nil {
		log.Err(err).Str("path", fl.path).Msg("failed to remove file")
	}
}
