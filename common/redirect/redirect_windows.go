package redirect

import (
	"os"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows"
)

var stderrFile *os.File
var originalStderr *os.File

func RedirectStderr(path string) error {
	// Create new file with desired permissions
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644)
	if err != nil {
		return err
	}

	// Call Windows API SetStdHandle
	err = windows.SetStdHandle(
		windows.STD_ERROR_HANDLE,
		windows.Handle(f.Fd()),
	)
	if err != nil {
		f.Close()
		os.Remove(path)
		return err
	}

	// Keep the file open
	stderrFile = f

	// Also set os.Stderr for Go's standard library
	originalStderr = os.Stderr
	os.Stderr = f

	return nil
}

// TODO: set original stderr back
func CloseStderr() error {
	if stderrFile != nil {
		err := stderrFile.Close()
		stderrFile = nil
		os.Stderr = originalStderr
		originalStderr = nil
		log.Logger = log.Output(os.Stderr)
		return err
	}
	return nil
}
