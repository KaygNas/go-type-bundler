package cmd

import (
	"fmt"
	gotypebundler "gotypebundler/pkg"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "gotypebundler",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var entry string
var output string
var types []string

func init() {
	generateCmd := &cobra.Command{
		Use:   "bundle -e <entry> -t <type1,type2> -o <output>",
		Short: "Bundle the specified package to a single Go file containing specific types.",
		Long:  "Bundle the specified package to a single Go file containing specific types.",
		Run: func(cmd *cobra.Command, args []string) {
			bundler := gotypebundler.New(&gotypebundler.Config{
				Entry:  entry,
				Types:  types,
				Output: output,
			})
			bundler.Bundle()
		},
	}
	generateCmd.Flags().StringVarP(&entry, "entry", "e", "", "Path to the directory containing the Go files to bundle")
	generateCmd.Flags().StringSliceVarP(&types, "types", "t", []string{}, "List of types to bundle")
	generateCmd.Flags().StringVarP(&output, "output", "o", "output.go", "Path to the output file")
	generateCmd.MarkFlagRequired("entry")
	rootCmd.AddCommand(generateCmd)
}
