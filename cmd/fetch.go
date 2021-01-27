/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fetcher/types"
	"fetcher/utils"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fetch(cmd)
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	fetchCmd.Flags().StringP("level", "l", "info", "log level")
	viper.BindPFlag("level", fetchCmd.Flags().Lookup("level"))

}

func fetch(cmd *cobra.Command) {

	var lvl logrus.Level
	switch viper.GetString("level") {
	case "debug":
		lvl = logrus.DebugLevel
	case "info":
		lvl = logrus.InfoLevel
	case "warning":
		lvl = logrus.WarnLevel
	case "error":
		lvl = logrus.ErrorLevel
	case "fatal":
		lvl = logrus.FatalLevel
	default:
		lvl = logrus.InfoLevel
	}
	fmt.Printf("Set log level: %+v\n", lvl)
	logrus.SetLevel(lvl)

	logrus.Infof("Config file: %v", cfgFile)

	var subscriptions []types.Subscription
	err := viper.UnmarshalKey("subscriptions", &subscriptions)
	if err != nil {
		logrus.Fatalf("Unmarshal subscriptions error: %v", err)
	}
	logrus.Debugf("subscriptions: %v", subscriptions)

	var customOuts []types.CustomOutbound
	viper.UnmarshalKey("custom_outbounds", &customOuts)

	templateFile := viper.GetString("template_file")
	startPort := viper.GetInt("start_port")
	logrus.Infof("template file: %v", templateFile)
	logrus.Infof("start port: %v", startPort)

	if len(subscriptions) == 0 {
		logrus.Warn("Cannot find subscrption url,not specify config file?")
	}
	configString, err := utils.Parse(subscriptions, templateFile, startPort, customOuts)
	if err != nil {
		logrus.Fatalf("Parse error: %v", err)
	}
	err = ioutil.WriteFile(viper.GetString("output_file"), []byte(configString), 0600)
	if err != nil {
		logrus.Fatalf("Write config file error: %v", err)
	}

}
