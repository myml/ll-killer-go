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

var RunFlag struct {
	Self  string
	Shell string
	Args  []string
}

func RunMain(ctx *cli.Context) error {
	args := []string{"ll-builder", "run"}
	args = append(args, ctx.Args().Slice()...)
	Exec(args...)
	return nil
}
func CreateRunCommand() *cli.Command {
	return &cli.Command{
		Name:        "run",
		Usage:       "启动容器",
		Description: "此命令执行ll-builder run，用于提供一致性体验。",
		Flags:       []cli.Flag{},
		Action:      RunMain,
	}
}
