package cmd

/**
 * Author: Matt Moran
 */

import (
	"os"
	"os/signal"

	"github.com/darkmattermatt/dumpdb/pkg/camelcase2underscore"

	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
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

var (
	signalInterrupt bool
	doneFile        *os.File
	skipFile        *os.File
	errFile         *os.File
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	l.FatalOnErr(err)
}

func init() {
	// initialize logger
	l.GetVerbosityWith(func() int {
		return l.INFO + v.GetInt("verbose") - v.GetInt("quiet")
	})

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.dumpdb.yaml)")
	rootCmd.PersistentFlags().CountP("verbose", "v", "verbosity. Set this flag multiple times for more verbosity")
	rootCmd.PersistentFlags().CountP("quiet", "q", "quiet. This is subtracted from the verbosity")

	v.BindPFlags(rootCmd.PersistentFlags())

	// listen for CTRL+C
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		signalInterrupt = true
		l.I("CTRL+C caught, stopping gracefully")
	}()
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
		l.FatalOnErr(err)

		// search config in home directory with name ".dumpdb" (without extension).
		v.AddConfigPath(home)
		v.SetConfigName(".dumpdb")
	}

	// if a config file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		l.V("Using config file:", v.ConfigFileUsed())
	}
}
