/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package main

import (
	"ll-killer/ptrace"
	"os"

	"github.com/spf13/cobra"
)

const (
	Usage = `ll-killer - Linglong Contrainer Utilities`
)
const MainCommandHelp = `
ll-killer是一个工具，旨在解决玲珑容器应用的构建问题。

使用ll-killer构建应用需要经历三个过程：创建项目、构建项目、提交文件

# 创建项目
创建项目阶段会创建linglong.yaml和一系列build-aux辅助构建脚本，以及apt.conf.d、sources.list.d等内容。
项目创建完成后会自动构建一次项目，用于初始化。
你可以使用以下命令创建一个项目：

<program> create --id appid

* 如果你是手动创建的linglong.yaml，需要手动执行一次ll-builder build以初始化环境，在未初始化的情况下，build命令中的严格模式不可用。

# 构建项目
构建项目阶段可以进入一个具有完全控制权限的根文件系统，你可以以root身份执行各种命令，包括apt安装命令。
你可以使用以下命令进入构建环境：

<program> build
可以在后面添加自定义构建命令:
<program> build -- 构建命令 参数...

* 构建环境的内容不会在环境退出后消失，你可以重复进入环境进行调整，如需清除内容请使用<program> clean命令。

# 提交内容
提交阶段用于将文件复制进玲珑容器，默认的build-aux/setup.sh构建命令已经实现了文件复制、快捷方式、图标文件的修复功能。
你可以使用以下命令进行提交：
<program> commit

* 此命令等同于 ll-builder build, 你可以在linglong.yaml中修改build构建指令，修改指令时请确保build-aux/setup.sh被调用。

# 测试应用
此命令等同于 ll-builder run，用于测试应用。
<program> run

# 其他操作
你可以直接使用ll-builder实现其他操作，如export导出layer和uab文件。

# 辅助构建脚本
build-aux文件夹中存储ll-killer使用的一些脚本工具，包括以下内容：

entrypoint.sh:          玲珑应用入口点
env.sh:                 辅助脚本环境配置脚本
ldd-check.sh:           检查容器内的缺失库，用于deb包未完整声明依赖的情况。
ldd-search.sh:          搜索缺失库所在的deb包名称，需要在ll-killer apt环境中使用。
						用法：ldd-search.sh <包括输入库文件名列表的文件> <找到的deb包> <失败的文件列表>
relink.sh               修复容器内的符号链接，玲珑不支持目录的相对符号链接。
						用法：relink.sh <符号链接>
setup-desktop.sh:       修复desktop文件中的Icon和Exec指令。
						用法：setup-desktop.sh: <desktop文件>
setup-filesystem.sh:    从构建环境中复制文件到$PREFIX。
setup-icon.sh:          将图标文件按尺寸放置到正确的位置，支持ico,png,jpeg/jpg,gif,bmp和svg文件。
						用法：setup-icon.sh: <图标文件>
						* 由于DDE的bug，仅搜索svg和png文件，因此所有类型的图片文件都将重命名为svg或png文件，即便它不是这两个格式。
setup.sh:               复制文件和创建入口点，并执行以上所有修复脚本。

`

func main() {
	if os.Getenv(kKillerDebug) != "" {
		GlobalFlag.Debug = true
	}
	err := SetupEnvVar()
	if err != nil {
		Debug("SetupEnvVar:", err)
	}
	cobra.EnableCommandSorting = false
	// log.SetFlags(0)
	app := cobra.Command{
		Use:   "ll-killer",
		Short: Usage,
		Long:  BuildHelpMessage(MainCommandHelp),
	}
	app.Flags().SortFlags = false
	app.InheritedFlags().SortFlags = false
	app.LocalFlags().SortFlags = false
	app.Flags().BoolVar(&GlobalFlag.Debug, "debug", GlobalFlag.Debug, "显示调试信息")
	app.AddCommand(CreateAPTCommand(),
		CreateBuildCommand(),
		CreateExecCommand(),
		CreateRunCommand(),
		CreateCommitCommand(),
		CreateCleanCommand(),
		CreateBuildAuxCommand(),
		CreateScriptCommand())
	if ptrace.IsSupported {
		app.AddCommand(CreatePtraceCommand())
	}
	if err := app.Execute(); err != nil {
		ExitWith(err)
	}

}
