package redirect

import (
	"os"

	"golang.org/x/sys/unix"
)

var (
	stderrFile     *os.File
	originalStderr *os.File
)

func RedirectStderr(path string) error {
	// Save the original stderr
	originalStderr = os.Stderr

	outputFile, err := os.Create(path)
	if err != nil {
		return err
	}
	err = unix.Dup2(int(outputFile.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		outputFile.Close()
		os.Remove(outputFile.Name())
		return err
	}
	stderrFile = outputFile

	return nil
}

func CloseStderr() error {
	if stderrFile != nil {
		stderrFile.Close()
		stderrFile = nil
	}

	// Restore original stderr
	if originalStderr != nil {
		unix.Dup2(int(originalStderr.Fd()), int(os.Stderr.Fd()))
		originalStderr = nil
	}

	return nil
}
