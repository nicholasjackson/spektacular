package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test command",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Get the current weather from weather.com for London")
		return nil
	},
}
