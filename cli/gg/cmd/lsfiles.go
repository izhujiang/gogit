/*
Copyright © 2022 Jiang Zhu <m.zhujiang@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"

	git "github.com/izhujiang/gogit/api"
	"github.com/spf13/cobra"
)

var (
	showStage bool
)

// lsFilesCmd represents the lsFiles command
var lsFilesCmd = &cobra.Command{
	Use:   "ls-files",
	Short: "Show information about files in the index and the working tree",
	Long: `This merges the file listing in the index with the actual working directory list, and shows different combinations of the two.
       -c, --cached
           Show cached files in the output (default)

       -d, --deleted
           Show deleted files in the output

       -m, --modified
           Show modified files in the output

       -o, --others
           Show other (i.e. untracked) files in the output

       -i, --ignored
           Show only ignored files in the output. When showing files in the index, print only those matched by an exclude pattern. When showing
           "other" files, show only those matched by an exclude pattern. Standard ignore rules are not automatically activated, therefore at least
           one of the --exclude* options is required.

       -s, --stage
           Show staged contents' mode bits, object name and stage number in the output.

       --directory
           If a whole directory is classified as "other", show just its name (with a trailing slash) and not its whole contents.

       --no-empty-directory
           Do not list empty directories. Has no effect without --directory.

       -u, --unmerged
           Show unmerged files in the output (forces --stage)

       -k, --killed
           Show files on the filesystem that need to be removed due to file/directory conflicts for checkout-index to succeed.
	   -z
           \0 line termination on output and do not quote filenames. See OUTPUT below for more information.

       --deduplicate
           When only filenames are shown, suppress duplicates that may come from having multiple stages during a merge, or giving --deleted and
           --modified option at the same time. When any of the -t, --unmerged, or --stage option is in use, this option has no effect.

       -x <pattern>, --exclude=<pattern>
           Skip untracked files matching pattern. Note that pattern is a shell wildcard pattern. See EXCLUDE PATTERNS below for more information.

       -X <file>, --exclude-from=<file>
           Read exclude patterns from <file>; 1 per line.

       --exclude-per-directory=<file>
           Read additional exclude patterns that apply only to the directory and its subdirectories in <file>.

       --exclude-standard
           Add the standard Git exclusions: .git/info/exclude, .gitignore in each directory, and the user’s global exclusion file.

       --error-unmatch
           If any <file> does not appear in the index, treat this as an error (return 1).

       --with-tree=<tree-ish>
           When using --error-unmatch to expand the user supplied <file> (i.e. path pattern) arguments to paths, pretend that paths which were
           removed in the index since the named <tree-ish> are still present. Using this option with -s or -u options does not make any sense.
`,
	Run: func(cmd *cobra.Command, args []string) {
		w := os.Stdout
		option := &git.LsFilesOption{Stage: true}

		git.LsFiles(w, option)

	},
}

func init() {
	lsFilesCmd.Flags().BoolVarP(&showStage, "stage", "s", false, "Show staged contents' mode bits, object name and stage number in the output.")
	rootCmd.AddCommand(lsFilesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsFilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsFilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
