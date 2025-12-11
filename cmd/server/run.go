//go:build server

package main

import (
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/logger"
	"github.com/5vnetwork/vx-core/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run",
	Long:  `run vx`,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("failed to get config, %v", err)
		return
	}

	l, err := logger.SetLog(config.Log)
	common.Must(err)
	defer l.Close()

	app, err := buildserver.NewX(config)
	if err != nil {
		log.Printf("failed to create server, %v", err)
		return
	}

	app.Run()
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
