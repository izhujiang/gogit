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
	"log"
	"os"

	git "github.com/izhujiang/gogit/api"
	"github.com/spf13/cobra"
)

var (
	writeToDatabase bool
	usingStdin      bool
)

// hashObjectCmd represents the hashObject command
var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("hashObject called")
		option := &git.HashObjectOption{
			ObjectType: "blob",
			Write:      writeToDatabase,
		}

		// TODO: read text from stdin or file specified by args[0]
		if len(args) > 0 {
			path := args[0]
			f, err := os.Open(path)
			defer f.Close()

			if err != nil {
				log.Fatal(err)
			}
			git.HashObject(f, os.Stdout, option)
		} else {
			if usingStdin {
				git.HashObject(os.Stdin, os.Stdout, option)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(hashObjectCmd)

	hashObjectCmd.Flags().BoolVarP(&writeToDatabase, "write", "w", false, "Actually write the object into the object database")
	hashObjectCmd.Flags().BoolVarP(&usingStdin, "stdin", "", false, "Read the object from standard input instead of from a file.")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hashObjectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hashObjectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
