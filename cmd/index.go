/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TheBigRoomXXL/tinysearch/tinysearch"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:  "index",
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdinArgs, err := stdinToArgs()
		if err != nil {
			return fmt.Errorf("fail to parse stdin: %w", err)
		}

		args = append(args, stdinArgs...)

		if len(args) == 0 {
			return fmt.Errorf("no argument found, you must input at least one seed url")
		}

		tinysearch.Index(args)
		return nil
	},
}

func stdinToArgs() ([]string, error) {
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("error parsing stdin: %s", err)
		}
		return strings.Fields(string(input)), nil
	}
	return []string{}, nil
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
