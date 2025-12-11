package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/5vnetwork/vx-core/common/sshhelper"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	OneKB = 1024
	TenKB = 10 * OneKB
	OneMB = 1024 * OneKB
	TenMB = 10 * OneMB
)

func InitZeroLog() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// file name and line number
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
}

func SetErrorLogLevel() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	// file name and line number
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func LoadEnvVariables(path string) error {
	env := godotenv.Load(path)
	if env != nil {
		return env
	}
	return nil
}

func GetTestSshClientRemote() (*sshhelper.Client, error) {
	privateKey, err := os.ReadFile(os.Getenv("SSH_TESTSERVER_SSH_KEY_PATH"))
	if err != nil {
		return nil, err
	}

	client, _, err := sshhelper.Dial(&sshhelper.DialConfig{
		Addr:                 fmt.Sprintf("%s:%d", os.Getenv("SSH_TESTSERVER_ADDRESS"), 22),
		User:                 os.Getenv("SSH_TESTSERVER_USERNAME"),
		PrivateKey:           privateKey,
		PrivateKeyPassphrase: os.Getenv("SSH_TESTSERVER_SSH_KEY_PASSPHRASE"),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetTestSshClientLocal() (*sshhelper.Client, error) {
	privateKey, err := os.ReadFile(os.Getenv("SSH_TESTSERVER_LOCAL_SSH_KEY_PATH"))
	if err != nil {
		return nil, err
	}

	client, _, err := sshhelper.Dial(&sshhelper.DialConfig{
		Addr:       fmt.Sprintf("%s:%d", os.Getenv("SSH_TESTSERVER_LOCAL_ADDRESS"), 22),
		User:       os.Getenv("SSH_TESTSERVER_LOCAL_USERNAME"),
		Password:   os.Getenv("SSH_TESTSERVER_LOCAL_PASSWORD"),
		PrivateKey: privateKey,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetTestSshClientLocalPassword() (*sshhelper.Client, error) {
	client, _, err := sshhelper.Dial(&sshhelper.DialConfig{
		Addr:     fmt.Sprintf("%s:%d", os.Getenv("SSH_TESTSERVER_LOCAL_ADDRESS"), 22),
		User:     os.Getenv("SSH_TESTSERVER_LOCAL_USERNAME"),
		Password: os.Getenv("SSH_TESTSERVER_LOCAL_PASSWORD"),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
