package main

import (
	"github.com/spf13/cobra"
	gc "github.com/untillpro/gochips"
)

func main() {

	var rootCmd = &cobra.Command{Use: "cder"}
	rootCmd.PersistentFlags().BoolVarP(&gc.IsVerbose, "verbose", "v", false, "Verbose output")

	// cmdCD
	{
		var cmdCD = &cobra.Command{
			Use:   "cd --repo <main-repo> --replace <repo2=<repo-to-replace]> [args]",
			Short: "Pull and build sources from given repos",
			Run:   runCmdCD,
		}

		cmdCD.PersistentFlags().StringVarP(&binaryName, "output", "o", "", "Output binary name")
		cmdCD.PersistentFlags().StringVarP(&workingDir, "working-dir", "w", ".", "Working directory")
		cmdCD.PersistentFlags().Int32VarP(&timeoutSec, "timeout", "t", 10, "Timeout")
		cmdCD.PersistentFlags().StringVarP(&mainRepo, "repo", "r", "", "Main repository")
		cmdCD.PersistentFlags().StringSliceVar(&argReplaced, "replace", []string{}, "Repositories to be replaced")
		cmdCD.PersistentFlags().StringSliceVar(&deployerEnv, "deployer-env", []string{}, "Deployer environment variable")
		cmdCD.MarkPersistentFlagRequired("repo")
		cmdCD.MarkPersistentFlagRequired("working-dir")

		rootCmd.AddCommand(cmdCD)
	}

	rootCmd.Execute()

}
