//go:build android

package x_android

import (
	"os"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

var stderrFile *os.File

func RedirectStderr(path string) error {
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

func UploadLog(message string) {
	log.Error().Msg(message)
}
