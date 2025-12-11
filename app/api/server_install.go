package api

import (
	"bytes"
	"fmt"

	"github.com/5vnetwork/vx-core/common/sshhelper"
	"github.com/rs/zerolog/log"
)

var HysteriaConfigPath = "/etc/hysteria/config.yaml"

func InstallHysteria(sshClient *sshhelper.Client, user string) error {
	o, err := sshClient.CombinedOutput(fmt.Sprintf("HYSTERIA_USER=%s bash -c \"$(curl -fsSL https://get.hy2.sh)\"", user), true)
	if err != nil {
		return fmt.Errorf("failed to run install-release.sh: %w. Output: %s", err, o)
	}
	log.Debug().Msgf("[bash <(curl -fsSL https://get.hy2.sh/)] succ output: %s", o)

	err = sshClient.Run("systemctl enable hysteria-server.service", true)
	if err != nil {
		return fmt.Errorf("failed to run enable hysteria-server: %w. Output: %s", err, o)
	}
	log.Debug().Msgf("[systemctl enable hysteria-server.service] succ output: %s", o)

	return nil
}

func UpdateHysteriaConfig(sshClient *sshhelper.Client, config []byte) error {
	err := sshClient.CopyContentToRemote(bytes.NewReader(config), "/tmp/hysteria.yaml", 644)
	if err != nil {
		return fmt.Errorf("failed to update hysteria config: %w", err)
	}

	o, err := sshClient.CombinedOutput(fmt.Sprintf("mv /tmp/hysteria.yaml %s", HysteriaConfigPath), true)
	if err != nil {
		return fmt.Errorf("failed to move hysteria.yaml: %w. Output: %s", err, o)
	}
	log.Debug().Msgf("[mv hysteria.yaml] output: %s", o)

	o, err = sshClient.CombinedOutput("systemctl restart hysteria-server.service", true)
	if err != nil {
		return fmt.Errorf("failed to restart hysteria-server.service: %w. Output: %s", err, o)
	}
	log.Debug().Msgf("[systemctl restart hysteria-server.service] succ output: %s", o)

	return nil
}

const xrayConfigPath = "/usr/local/etc/xray/config.json"

func InstallXray(sshClient *sshhelper.Client, user string) error {
	// check if xray is installed, if so, uninstall it
	existed, err := sshClient.CommandExists("xray", true)
	if err != nil {
		return fmt.Errorf("failed to check if xray is installed: %w", err)
	}
	if existed {
		o, err := sshClient.CombinedOutput("bash -c \"$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)\" @ remove", true)
		if err != nil {
			return fmt.Errorf("failed to uninstall xray: %w. Output: %s", err, o)
		}
		log.Debug().Msgf("[xray uninstall] output: %s", o)
	}

	fmt.Println("downloading install-release.sh")
	o, err := sshClient.CombinedOutput("bash -c \"$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)\" @ install -u "+user, true)
	if err != nil {
		return fmt.Errorf("failed to run download install-release.sh: %w. Output: %s", err, o)
	}
	log.Debug().Msg("wget script output: \n" + o)

	// o, err = sshClient.CombinedOutput("bash install-release.sh", true)
	// if err != nil {
	// 	return fmt.Errorf("failed to run install-release.sh: %w. Output: %s.", err, o)
	// }
	// log.Debug().Msgf("[bash install-release.sh] succ output: %s", o)

	o, err = sshClient.CombinedOutput("systemctl enable xray", true)
	if err != nil {
		return fmt.Errorf("failed to run enable xray: %w. Output: %s", err, o)
	}
	log.Debug().Msgf("[systemctl enable xray] output: %s", o)

	return nil
}

func UpdateXrayConfig(sshClient *sshhelper.Client, config []byte) error {
	err := sshClient.CopyContentToRemote(bytes.NewReader(config), "/tmp/config.json", 644)
	if err != nil {
		return fmt.Errorf("failed to update xray config: %w", err)
	}

	o, err := sshClient.CombinedOutput(fmt.Sprintf("mv /tmp/config.json %s", xrayConfigPath), true)
	if err != nil {
		return fmt.Errorf("failed to move config.json: %w. Output: %s", err, o)
	}
	log.Debug().Msgf("[mv config.json] output: %s", o)

	o, err = sshClient.CombinedOutput("systemctl restart xray", true)
	if err != nil {
		return fmt.Errorf("failed to restart xray: %w. Output: %s", err, o)
	}
	log.Debug().Msgf("[systemctl restart xray] succ output: %s", o)

	return nil
}

func InstallVX(sshClient *sshhelper.Client, user string) error {
	existed, err := sshClient.CommandExists("vx", true)
	if err != nil {
		return fmt.Errorf("failed to check if vx is installed: %w", err)
	}
	if existed {
		// o, err := sshClient.CombinedOutput("bash -c \"$(curl -L https://github.com/5vnetwork/vx-install/raw/main/install-vx.sh)\" @ remove", true)
		// if err != nil {
		// 	return fmt.Errorf("failed to uninstall vx: %w. Output: %s", err, o)
		// }
		// log.Debug().Msgf("[vx uninstall] output: %s", o)
		return nil
	}

	fmt.Println("downloading install-vx.sh")
	o, err := sshClient.CombinedOutput("bash -c \"$(curl -L https://github.com/5vnetwork/vx-install/raw/main/install-vx.sh)\" @ install -u "+user, true)
	if err != nil {
		return fmt.Errorf("failed to run download install-vx.sh: %w. Output: %s", err, o)
	}
	log.Debug().Msg("wget script output: \n" + o)

	o, err = sshClient.CombinedOutput("systemctl enable vx", true)
	if err != nil {
		return fmt.Errorf("failed to run enable vx: %w. Output: %s", err, o)
	}
	log.Debug().Msgf("[systemctl enable vx] output: %s", o)

	return nil
}

func UninstallVX(sshClient *sshhelper.Client) error {
	o, err := sshClient.CombinedOutput("bash -c \"$(curl -L https://github.com/5vnetwork/vx-install/raw/main/install-vx.sh)\" @ remove", true)
	if err != nil {
		return fmt.Errorf("failed to uninstall vx: %w. Output: %s", err, o)
	}
	return nil
}

func UpdateVX(sshClient *sshhelper.Client, user string) error {
	o, err := sshClient.CombinedOutput("bash -c \"$(curl -L https://github.com/5vnetwork/vx-install/raw/main/install-vx.sh)\" @ install -u "+user, true)
	if err != nil {
		return fmt.Errorf("failed to run download install-vx.sh: %w. Output: %s", err, o)
	}
	return nil
}
