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

var ScriptFlag struct {
}

const ScriptCommandHelp = `
此命令为命令提供KILLER_EXEC环境变量，指向ll-killer二进制的绝对路径，确保构建脚本能够找到ll-killer。
用法：
<program> script -- <构建脚本> [参数...]

此命令等同于: 
KILLER_EXEC=<program> <构建脚本> [参数...]
`

func ScriptMain(ctx *cli.Context) error {

	return ExecRaw(ctx.Args().Slice()...)
}

func CreateScriptCommand() *cli.Command {
	return &cli.Command{
		Name:        "script",
		Description: BuildHelpMessage(ScriptCommandHelp),
		Usage:       "执行自定义构建流程",
		Flags:       []cli.Flag{},
		Action:      ScriptMain,
	}
}
