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
	prefix string
)

// readTreeCmd represents the readTree command
var readTreeCmd = &cobra.Command{
	Use:   "read-tree",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			w := os.Stdout
			id := args[0]
			option := &git.ReadTreeOption{
				Prefix: prefix,
			}

			git.ReadTree(w, id, option)

		}

	},
}

func init() {
	// Keep the current index contents, and read the contents of the named tree-ish under the directory at
	// <prefix>. The command will refuse to overwrite entries that already existed in the original index file.
	readTreeCmd.Flags().StringVar(&prefix, "prefix", "", "Keep the current index contents, and read the contents of the named tree-ish under the directory at <prefix>")
	rootCmd.AddCommand(readTreeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// readTreeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// readTreeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
