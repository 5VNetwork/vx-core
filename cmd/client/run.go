//go:build !android

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/common/redirect"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run vx core",
	Long:  `run vx core`,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("failed to get config, %v", err)
		return
	}

	if config.RedirectStdErr != "" {
		err := redirect.RedirectStderr(config.RedirectStdErr)
		if err != nil {
			log.Printf("failed to redirect stderr, %v", err)
			return
		}
	}

	server, err := buildclient.NewX(config)
	if err != nil {
		log.Print(err)
		return
	}

	err = server.Start()
	if err != nil {
		log.Print(err)
		return
	}
	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		<-osSignals
	}
	err = server.Close()
	if err != nil {
		log.Print(err)
	}
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
