package main

import (
	"log"

	_ "embed"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/eval/runner"
	"github.com/spf13/cobra"
)

var (
	runFlag string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "eval",
		Short: "AI evaluation runner",
		Long:  "A tool for running AI evaluation functions",
	}

	var runCmd = &cobra.Command{
		Use:   "run [path]",
		Short: "Run evaluation functions",
		Long:  "Run evaluation functions. If no path is provided, defaults to './...'",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Default to "./..." if no path provided
			paths := []string{"./..."}
			if len(args) > 0 {
				paths = args
			}

			funcs, err := runner.FindEvals(paths...)
			if err != nil {
				log.Fatal(err)
			}

			// Filter by function name if -run flag is provided
			if runFlag != "" {
				funcs = runner.FilterEvalsByName(funcs, runFlag)
			}

			err = runner.RunEvalsByModule(funcs)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	runCmd.Flags().StringVarP(&runFlag, "run", "r", "", "Run only eval function whose name exactly matches the provided name")

	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
