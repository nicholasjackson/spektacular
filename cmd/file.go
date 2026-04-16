package cmd

import (
	"fmt"
	"io"

	"github.com/jumppad-labs/spektacular/internal/output"
	"github.com/jumppad-labs/spektacular/internal/store"
	"github.com/spf13/cobra"
)

var specFileCmd = &cobra.Command{
	Use:   "file",
	Short: "Read and write files in the spec store",
}

var specFileWriteCmd = &cobra.Command{
	Use:   "write <path>",
	Short: "Write stdin to a file in the spec store",
	Args:  cobra.ExactArgs(1),
	RunE:  runSpecFileWrite,
}

var specFileReadCmd = &cobra.Command{
	Use:   "read <path>",
	Short: "Read a file from the spec store and write it to stdout",
	Args:  cobra.ExactArgs(1),
	RunE:  runSpecFileRead,
}

var specFileDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a file from the spec store",
	Args:  cobra.ExactArgs(1),
	RunE:  runSpecFileDelete,
}

var specFileListCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List files in the spec store",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSpecFileList,
}

func runSpecFileWrite(cmd *cobra.Command, args []string) error {
	dataDir, err := dataDir()
	if err != nil {
		return err
	}
	content, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}
	return store.NewFileStore(dataDir).Write(args[0], content)
}

func runSpecFileRead(cmd *cobra.Command, args []string) error {
	dataDir, err := dataDir()
	if err != nil {
		return err
	}
	content, err := store.NewFileStore(dataDir).Read(args[0])
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(content)
	return err
}

func runSpecFileDelete(_ *cobra.Command, args []string) error {
	dataDir, err := dataDir()
	if err != nil {
		return err
	}
	return store.NewFileStore(dataDir).Delete(args[0])
}

func runSpecFileList(cmd *cobra.Command, args []string) error {
	dataDir, err := dataDir()
	if err != nil {
		return err
	}
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	entries, err := store.NewFileStore(dataDir).List(path)
	if err != nil {
		return err
	}
	return output.Write(cmd.OutOrStdout(), map[string]any{"files": entries}, "")
}

func init() {
	specFileCmd.AddCommand(specFileWriteCmd, specFileReadCmd, specFileDeleteCmd, specFileListCmd)
	specCmd.AddCommand(specFileCmd)
}
