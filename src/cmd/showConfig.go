package cmd

import (
	"github.com/felbinger/GBM/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

// showConfigCmd represents the showConfig command
var showConfigCmd = &cobra.Command{
	Use:   "showConfig",
	Short: "Show configuration",
	Run: func(cmd *cobra.Command, args []string) {
		secure, err := strconv.ParseBool((cmd.Flag("secure").Value).String())
		if err != nil {
			log.Fatal("Unable to parse secure parameter. Exiting...")
		}
		utils.ShowConfiguration(Configuration, secure)
	},
}

func init() {
	showConfigCmd.PersistentFlags().Bool("secure", false, "Secure configuration (hide passwords)")
	rootCmd.AddCommand(showConfigCmd)
}
