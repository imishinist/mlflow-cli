package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "mlflow-cli",
	Short: "MLflow Tracking CLI Tool",
	Long: `A command line tool for MLflow tracking operations.
Supports logging parameters, metrics, and artifacts to MLflow tracking server.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().String("tracking-uri", "", "MLflow tracking URI (overrides MLFLOW_TRACKING_URI)")
	rootCmd.PersistentFlags().String("experiment-id", "", "Experiment ID (overrides MLFLOW_EXPERIMENT_ID)")
	viper.BindPFlag("tracking_uri", rootCmd.PersistentFlags().Lookup("tracking-uri"))
	viper.BindPFlag("experiment_id", rootCmd.PersistentFlags().Lookup("experiment-id"))
}

func initConfig() {
	// Environment variables
	viper.SetEnvPrefix("MLFLOW")
	viper.AutomaticEnv()

	// Also bind Databricks environment variables
	viper.BindEnv("databricks_host", "DATABRICKS_HOST")
	viper.BindEnv("databricks_token", "DATABRICKS_TOKEN")

	// Set defaults
	viper.SetDefault("tracking_uri", "http://localhost:5000")
	viper.SetDefault("time_resolution", "1m")
	viper.SetDefault("time_alignment", "floor")
	viper.SetDefault("step_mode", "auto")
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
