/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package _commit

import (
	"ll-killer/utils"

	"github.com/spf13/cobra"
)

var CommitFlag struct {
	Self  string
	Shell string
	Args  []string
}

func CommitMain(cmd *cobra.Command, args []string) error {
	args = append([]string{"ll-builder", "build"}, args...)
	utils.Exec(args...)
	return nil
}
func CreateCommitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "commit",
		Short:              "提交构建内容",
		Long:               "此命令执行ll-builder build，用于提供一致性体验。",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			utils.ExitWith(CommitMain(cmd, args))
		},
	}

	// 由于原命令没有标志，这里不添加 Flags
	return cmd
}
