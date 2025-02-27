/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package main

import (
	"github.com/urfave/cli/v2"
)

var CommitFlag struct {
	Self  string
	Shell string
	Args  []string
}

func CommitMain(ctx *cli.Context) error {
	args := []string{"ll-builder", "build"}
	args = append(args, ctx.Args().Slice()...)
	Exec(args...)
	return nil
}
func CreateCommitCommand() *cli.Command {
	return &cli.Command{
		Name:        "commit",
		Usage:       "提交构建内容",
		Description: "此命令执行ll-builder build，用于提供一致性体验。",
		Flags:       []cli.Flag{},
		Action:      CommitMain,
	}
}
