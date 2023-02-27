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
	"fmt"
	"strings"

	git "github.com/izhujiang/gogit/api"
	"github.com/spf13/cobra"
)

var (
	replace   bool
	add       bool
	remove    bool
	cacheinfo string
)

// updateIndexCmd represents the updateIndex command
var updateIndexCmd = &cobra.Command{
	Use:   "update-index",
	Short: "Register file contents in the working tree to the index",
	Long: `Modifies the index. Each file mentioned is updated into the index and any unmerged or needs updating state is cleared.
       See also git-add(1) for a more user-friendly way to do some of the most common operations on the index.
`,
	Run: func(cmd *cobra.Command, args []string) {
		option := &git.UpdateIndexOption{}
		option.Args = make(map[string]string)

		// TODO: make flags MutuallyExclusive
		if add {
			option.Op = "add"
		} else if remove {
			option.Op = "remove"
		} else {
			option.Op = "replace"
		}

		if len(args) == 0 && cacheinfo == "" {
			fmt.Println("usage: git update-index [--add] [--remove | --force-remove] [--replace] [(--cacheinfo <mode>,<object>,<file>)...] [--] [<file>...]")
			return
		}

		if cacheinfo != "" {
			cacheItems := strings.Split(cacheinfo, ",")

			if len(cacheItems) < 3 {
				fmt.Println("option 'cacheinfo' expects <mode>,<sha1>,<path>")
				return
			}
			option.Args["mode"] = cacheItems[0]
			option.Args["oid"] = cacheItems[1]
			option.Args["file"] = cacheItems[2]
		}

		if len(args) > 0 {
			option.Path = args[0]
		}

		git.UpdateIndex(option)
	},
}

func init() {

	updateIndexCmd.Flags().BoolVar(&replace, "replace", true, `By default, when a file path exists in the index, git update-index refuses an attempt to add path/file. Similarly if a file path/file exists, a file path cannot be added. With --replace flag, existing entries that conflict with the entry being added are automatically removed with warning messages.`)
	updateIndexCmd.Flags().BoolVar(&add, "add", false, "If a specified file isn’t in the index already then it’s added. Default behaviour is to ignore new files")
	updateIndexCmd.Flags().BoolVar(&remove, "remove", false, "If a specified file is in the index but is missing then it’s removed. Default behavior is to ignore removed file.")
	updateIndexCmd.MarkFlagsMutuallyExclusive("replace", "add", "remove")
	// updateIndexCmd.MarkFlagsMutuallyExclusive("add", "remove")

	updateIndexCmd.Flags().StringVar(&cacheinfo, "cacheinfo", "", "Directly insert the specified info into the index.")

	rootCmd.AddCommand(updateIndexCmd)

}
