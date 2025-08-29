package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	verbose    bool
	jsonOutput bool
)

var rootCmd = &cobra.Command{
	Use:   "godisplay",
	Short: "Minimal display management for macOS",
	Long: `godisplay is a lightweight CLI tool for managing display resolutions on macOS.
It provides direct access to display configuration without GUI overhead.

This tool requires macOS 10.15+ and will not work in sandboxed environments.`,
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.godisplay.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"verbose output")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false,
		"output in JSON format")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".godisplay")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GODISPLAY")

	// Set defaults
	viper.SetDefault("default_refresh_rate", 60)
	viper.SetDefault("prefer_hidpi", true)
	viper.SetDefault("safe_mode", true) // Prevent setting resolutions below 800x600

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintf(os.Stderr, "Using config file: %s", viper.ConfigFileUsed())
	}
}
