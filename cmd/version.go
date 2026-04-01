package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X ...".
var Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the rei version",
	Long:  `Print the current version of the rei CLI.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("rei version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
