/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package _pack

import (
	"ll-killer/layer"
	"ll-killer/utils"
	"os"

	"github.com/spf13/cobra"
)

const PackCommandHelp = `将文件夹打包为layer,此命令直接使用mkfs.erofs对玲珑的输出目录进行打包，从而无需ll-builder提交到ostree，减少不必要的文件复制。
* 使用ll-builder build --skip-commit-output跳过commit来得到构建脚本的原始输出linglong/output/_build/。
* 原始输出未对快捷方式/systemd/dbus服务添加'll-cli run {APPID} -- '前缀，需要二次处理。

本功能的终极目的是彻底消除~/.cache/linglong-builder下的应用缓存，让ll-builder run直接运行在linglong/output/binary上，保护磁盘寿命。

目前实际可操作的流程如下：

* 手动处理linglong/output/_build/，并创建类似linglong/output/binary的目录结构。
* 挂载已有的layer，复制layer内容并修改，最后重新打包。
* 挂载已有的layer，使用overlayfs叠加layer目录，在合并目录中修改，最后对合并目录重新打包。
* 挂载已有的高版本layer，使用lz4hc算法重新打包，避免旧版本erofs不支持zstd压缩算法，旧版玲珑无法安装的问题。

`

var Flag layer.PackOption

func PackMain(cmd *cobra.Command, args []string) error {
	utils.Debug("PackMain", args)
	Flag.Args = args
	return layer.Pack(&Flag)
}

func CreatePackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pack <源目录> [layer文件名] [flags] -- [mkfs.erofs选项]",
		Short:         "将文件夹打包为layer。",
		Long:          PackCommandHelp,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			Flag.Source = args[0]
			utils.ExitWith(PackMain(cmd, args[1:]))
		},
	}
	cmd.Flags().IntVarP(&Flag.BlockSize, "block-size", "b", 4096, "块大小")
	cmd.Flags().StringVarP(&Flag.Target, "output", "o", "", "输出文件名")
	cmd.Flags().StringVar(&Flag.ExecPath, "exec", layer.MkfsErofs, "指定mkfs.erofs命令位置")
	cmd.Flags().IntVarP(&Flag.Uid, "force-uid", "U", os.Getuid(), "文件Uid,-1为不更改")
	cmd.Flags().IntVarP(&Flag.Gid, "force-gid", "G", os.Getegid(), "文件Gid,-1为不更改")
	cmd.Flags().StringVarP(&Flag.Compressor, "compressor", "z", "lz4hc", "压缩算法，请查看mkfs.erofs帮助")
	return cmd
}
