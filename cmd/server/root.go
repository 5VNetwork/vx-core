//go:build !android

package main

import (
	"os"

	"github.com/5vnetwork/vx-core/common"
	"github.com/spf13/cobra"
)

var CfgFile string
var logLevel string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "vx",
	Short:   "vx is a proxy tool",
	Version: common.Version,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(readInConfig)
	rootCmd.PersistentFlags().StringVar(&CfgFile, "config", "config.json", "config file (default is ./config.json)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Set custom version template
	// rootCmd.SetVersionTemplate(fmt.Sprintf("vx version %s\n", Version))
}

// initConfig reads in config file and ENV variables if set.
// func readInConfig() {
// 	if cfgFile != "" {
// 		// Use config file from the flag.
// 		viper.SetConfigFile(cfgFile)
// 	} else {
// 		// Search config in home directory with name ".github.com/5vnetwork/vx-core" (without extension).
// 		viper.AddConfigPath(".")
// 		viper.SetConfigType("json")
// 		viper.SetConfigName("config")
// 	}

// 	viper.AutomaticEnv() // read in environment variables that match

// 	// If a config file is found, read it in.
// 	if err := viper.ReadInConfig(); err == nil {
// 		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
// 	}
// }
