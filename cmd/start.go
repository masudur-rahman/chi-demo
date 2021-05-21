package cmd

import (
	"github.com/masudur-rahman/chi-demo/api"
	"github.com/spf13/cobra"
)

var host string
var port string

var startApp = &cobra.Command{
	Use:   "start",
	Short: "Start the app",
	Long:  "This starts the apiserver",
	Run: func(cmd *cobra.Command, args []string) {
		api.AssignValues(host, port)
		api.StartTheApp()
	},
}

func init() {
	startApp.PersistentFlags().StringVar(&host, "host", "0.0.0.0", "host address for the server")
	startApp.PersistentFlags().StringVarP(&port, "port", "p", "4000", "port number for the server")

	rootCmd.AddCommand(startApp)
}
