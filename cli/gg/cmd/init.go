/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	git "github.com/izhujiang/gogit/api"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create an empty Git repository or reinitialize an existing one",
	Long: `This command creates an empty Git repository - basically a .git directory with subdirectories for objects, refs/heads, refs/tags, and template files.
	   If the $GIT_DIR environment variable is set then it specifies a path to use instead of ./.git for the base of the repository.

       If the object storage directory is specified via the $GIT_OBJECT_DIRECTORY environment variable then the sha1 directories are created underneath - otherwise the default $GIT_DIR/objects directory is used.

       Running git init in an existing repository is safe. It will not overwrite things that are already there. The primary reason for rerunning git init is to pick up newly added templates (or to move the repository to another place if --separate-git-dir is given).
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			git.Init(os.Stdout, args[0])

		} else {
			git.Init(os.Stdout, "")
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

}
