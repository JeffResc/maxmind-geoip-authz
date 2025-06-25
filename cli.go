package main

import (
	"log"

	cfg "github.com/jeffresc/maxmind-geoip-authz/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "geoip",
	Short: "MaxMind GeoIP authorization tool",
}

func init() {
	rootCmd.AddCommand(serveCmd)
	updateCmd.AddCommand(updateDBCmd)
	rootCmd.AddCommand(updateCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serve()
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update GeoIP resources",
}

var updateDBCmd = &cobra.Command{
	Use:   "database",
	Short: "Update the GeoIP database",
	Run: func(cmd *cobra.Command, args []string) {
		config = cfg.Load("config.yaml")
		_, licenseKey = cfg.LoadMaxMindCredentials(
			config.MaxMindAccountIDFile,
			config.MaxMindLicenseKeyFile,
		)
		downloadGeoIPDBIfUpdated()
	},
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
