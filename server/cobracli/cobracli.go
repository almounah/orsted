package cobracli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"orsted/server/core"
)

var rootCmd = &cobra.Command{
	Use:   "orsted-server",
	Short: "orsted-server, the C2 of the orsted server",
	Run: func(cmd *cobra.Command, args []string) {
		// This function will be called when no subcommands are provided
		fmt.Println("Use --help to see available commands.")
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the server",
	Long:  `The greet command prints a friendly greeting message to the user.`,
	Run: func(cmd *cobra.Command, args []string) {
		core.StartServerGRPC("0.0.0.0", 50051)
	},
}

func Init() {
    rootCmd.AddCommand(runCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
    Init()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
