package cmd

/*
Copyright © 2019 Ivan Webber <ivan.deacon.webber@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ivanthewebber/cs372-project/rot13"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	input  string
	amount int
	unrot  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rot [string to rotate]",
	Short: "Rot is a Caesar cipher tool for English useful for obsfucating questionable content.",
	Long: `Rot (i.e. rotate) is a Ceasar cipher tool that by default uses ROT13.

ROT13 is used in online forums as a means of hiding spoilers, punchlines, puzzle solutions, and offensive materials from the casual glance. ROT13 has inspired a variety of letter and word games on-line, and is frequently mentioned in newsgroup conversations.
    - https://en.wikipedia.org/wiki/ROT13
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if unrot {
			amount = -amount
		}

		io.Copy(os.Stdout, rot13.NewReader(strings.NewReader(strings.Join(args, " ")), amount))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	viper.AutomaticEnv() // read in environment variables that match

	rootCmd.Flags().IntVarP(&amount, "amount", "a", 13, "amount to rotate alphabet letters")
	rootCmd.Flags().BoolVarP(&unrot, "unrot", "u", false, "undoes rotation")
}
