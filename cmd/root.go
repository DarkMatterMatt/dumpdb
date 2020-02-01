/*
 * Author: Matt Moran
 */
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/darkmattermatt/dumpdb/pkg"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var v *viper.Viper

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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.DumpDB.yaml)")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	// read in environment variables that match
	v = viper.NewWithOptions(viper.EnvKeyReplacer(pkg.NewCamelcaseToUnderscoreReplacer()))
	v.SetEnvPrefix("ddb")
	v.AutomaticEnv()

	if cfgFile != "" {
		// use config file from the flag
		v.SetConfigFile(cfgFile)
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
