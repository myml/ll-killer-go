/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package layer

import (
	"ll-killer/utils"

	"github.com/spf13/cobra"
)

var PackFlag struct {
	Self  string
	Shell string
	Args  []string
}

func PackMain(cmd *cobra.Command, args []string) error {
	args = append([]string{"ll-builder", "run"}, args...)
	// Exec(args...)
	return nil
}

func CreatePackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pack",
		Short: "打包layer",
		Run: func(cmd *cobra.Command, args []string) {
			utils.ExitWith(PackMain(cmd, args))
		},
	}

	return cmd
}
