package cmd

/**
 * Author: Matt Moran
 */

import (
	"fmt"
	"os"

	"github.com/darkmattermatt/dumpdb/pkg/camelcase2underscore"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var v = viper.NewWithOptions(viper.EnvKeyReplacer(camelcase2underscore.NewCamelcase2UnderscoreReplacer()))

// the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "DumpDB",
	Short: "DumpDB imports credential dumps into a database to improve search performance.",
	Long:  "",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.dumpdb.yaml)")
	rootCmd.PersistentFlags().CountP("verbose", "v", "verbosity. Set this flag multiple times for more verbosity")
	rootCmd.PersistentFlags().CountP("quiet", "q", "quiet. This is subtracted from the verbosity")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	// read in environment variables that match
	v.SetEnvPrefix("ddb")
	v.AutomaticEnv()

	if v.GetString("config") != "" {
		// use config file from the flag
		v.SetConfigFile(v.GetString("config"))
	} else {
		// find home directory
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// search config in home directory with name ".dumpdb" (without extension).
		v.AddConfigPath(home)
		v.SetConfigName(".dumpdb")
	}

	// if a config file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", v.ConfigFileUsed())
	}
}
