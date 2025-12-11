//go:build darwin

package x_darwin

import (
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

var stderrFile *os.File

func RedirectStderr(path string) error {
	// if the file already exists, rename it
	if _, err := os.Stat(path); err == nil {
		// rename the file to a timestamped file
		timestamp := time.Now().Format("20060102150405")
		dir := filepath.Dir(path)
		os.Rename(path, filepath.Join(dir, timestamp+".txt"))
	}

	outputFile, err := os.Create(path)
	if err != nil {
		return err
	}
	err = outputFile.Chown(os.Getuid(), os.Getgid())
	if err != nil {
		outputFile.Close()
		os.Remove(outputFile.Name())
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
