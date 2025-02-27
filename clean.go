/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package main

import (
	"github.com/spf13/cobra"
)

var CleanFlag struct {
	FileSystem bool
	APT        bool
}

func CleanMain(cmd *cobra.Command, args []string) error {
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

func CreateCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "清除构建内容",
		RunE:  CleanMain,
	}

	cmd.Flags().BoolVar(&CleanFlag.FileSystem, "filesystem", true, "清除容器文件系统")
	cmd.Flags().BoolVar(&CleanFlag.APT, "apt", false, "清除APT缓存")
	return cmd
}
