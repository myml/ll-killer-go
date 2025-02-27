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

var CleanFlag struct {
	FileSystem bool
	APT        bool
}

func CleanMain(ctx *cli.Context) error {
	if CleanFlag.FileSystem {
		err := RunCommand("rm", "-rf", FileSystemDir)
		if err != nil {
			return err
		}
	}

	if CleanFlag.APT {
		err := RunCommand("rm", "-rf", AptDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateCleanCommand() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "清除构建内容",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "filesystem",
				Usage:       "清除容器文件系统",
				Value:       true,
				Destination: &CleanFlag.FileSystem,
			},
			&cli.BoolFlag{
				Name:        "apt",
				Usage:       "清除APT缓存",
				Value:       false,
				Destination: &CleanFlag.APT,
			},
		},
		Action: CleanMain,
	}
}
