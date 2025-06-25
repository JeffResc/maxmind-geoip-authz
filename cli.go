package main

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "geoip"}

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

var updateCmd = &cobra.Command{Use: "update"}

var updateDBCmd = &cobra.Command{
	Use:   "database",
	Short: "Update the GeoIP database",
	Run: func(cmd *cobra.Command, args []string) {
		config = LoadConfig("config.yaml")
		accountID, licenseKey = LoadMaxMindCredentials(
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
