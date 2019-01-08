package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"clarin/shib-aagregator/src/server"
	"clarin/shib-aagregator/src/logger"
)

var log_level string
var port int
var sp_entity_id string
var simulate bool
var aagregator_url string
var aagregator_path string

var ServerCmd = &cobra.Command{
	Use:   "aagregator",
	Short: "Shibboleth attribute aggregator cli",
	Long: `Control interface for the shibboleth attribute aggregator.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of docker-clarin",
	Long:  `All software has versions. This is docker-clarin's.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Shibboleth test server v%s by CLARIN ERIC\n", "1.0.0-beta")
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start server",
	Long: `Start server`,
	Run: func(cmd *cobra.Command, args []string) {
		logPtr := logger.NewLogger(log_level)
		server.StartServerAndblock(logPtr, port, sp_entity_id, simulate, aagregator_url, aagregator_path)
	},
}

func Execute() {
	startCmd.Flags().IntVarP(&port, "port", "p", 8080, "Base port to run the server on")
	startCmd.Flags().StringVarP(&sp_entity_id, "spEntityId", "i", "sp.clarin.eu", "Specify the entity id for this SP.")
	startCmd.Flags().BoolVarP(&simulate, "simulate", "s", false, "Simulate statistics submission. This mode will not send any data to the remote aagregator endpoint.")
	startCmd.Flags().StringVarP(&aagregator_url, "aagregator_url", "U", "https://clarin-aa.ms.mff.cuni.cz", "Base URL for the aagregator endpoint.")
	startCmd.Flags().StringVarP(&aagregator_path, "aagregator_path", "P", "/aaggreg/v1/got", "Path under the base URL for the aagregator endpoint.")

	ServerCmd.PersistentFlags().StringVarP(&log_level, "log_level", "v", logger.LevelToString(logger.GetDefaultLogLevel()), "Log level, supported values: TRACE, DEBUG, INFO, WARN and ERROR.")
	ServerCmd.AddCommand(versionCmd)
	ServerCmd.AddCommand(startCmd)
	ServerCmd.Execute()
}