package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	XUDPBaseKey = "xray.xudp.basekey"
)

type EnvFlag struct {
	Name    string
	AltName string
}

func NewEnvFlag(name string) EnvFlag {
	return EnvFlag{
		Name:    name,
		AltName: NormalizeEnvName(name),
	}
}

func (f EnvFlag) GetValue(defaultValue func() string) string {
	if v, found := os.LookupEnv(f.Name); found {
		return v
	}
	if len(f.AltName) > 0 {
		if v, found := os.LookupEnv(f.AltName); found {
			return v
		}
	}

	return defaultValue()
}

func (f EnvFlag) GetValueAsInt(defaultValue int) int {
	useDefaultValue := false
	s := f.GetValue(func() string {
		useDefaultValue = true
		return ""
	})
	if useDefaultValue {
		return defaultValue
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int(v)
}

func NormalizeEnvName(name string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(name)), ".", "_")
}

func getExecutableDir() string {
	exec, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exec)
}

// GetConfigurationPath reads "x.location.config" environment variable,
// and returns the path to the configuration file. if not set, it returns
// the default configuration path, which is the same directory as the executable.
func GetConfigurationPath() string {
	if runtime.GOOS == "linux" {
		return "/usr/local/etc/github.com/5vnetwork/vx-core/config.json"
	} else {
		return filepath.Join(getExecutableDir(), "config.json")
	}
}

func GetAssetLocation(filename string) string {
	const assetPathEnvKey = "x.location.asset"
	configPath := NewEnvFlag(assetPathEnvKey).GetValue(getExecutableDir)
	return filepath.Join(configPath, filename)
}
