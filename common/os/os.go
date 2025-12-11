package os

import (
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sys/cpu"
)

func GetExecutableDir() string {
	exec, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exec)
}

var (
	HasGCMAsmAMD64 = cpu.X86.HasAES && cpu.X86.HasPCLMULQDQ
	HasGCMAsmARM64 = cpu.ARM64.HasAES && cpu.ARM64.HasPMULL
	// Keep in sync with crypto/aes/cipher_s390x.go.
	HasGCMAsmS390X = cpu.S390X.HasAES && cpu.S390X.HasAESCBC && cpu.S390X.HasAESCTR &&
		(cpu.S390X.HasGHASH || cpu.S390X.HasAESGCM)

	HasAESGCMHardwareSupport = runtime.GOARCH == "amd64" && HasGCMAsmAMD64 ||
		runtime.GOARCH == "arm64" && HasGCMAsmARM64 ||
		runtime.GOARCH == "s390x" && HasGCMAsmS390X
)
