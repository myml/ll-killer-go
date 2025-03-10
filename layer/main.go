/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package layer

import "github.com/spf13/cobra"

func CreateLayerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "layer",
		Short: "layer 打包/挂载/解压工具",
		Long:  "独立于玲珑的layer管理器，提供高效的功能支持，需要安装erofs-utils。",
	}
	cmd.AddCommand(CreatePackCommand())
	return cmd
}
