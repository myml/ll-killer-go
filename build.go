/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package main

import (
	"encoding/base64"
	"fmt"
	"ll-killer/ptrace"
	"log"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/moby/sys/reexec"
	"github.com/spf13/cobra"
)

var BuildFlag struct {
	RootFS      string
	TmpRootFS   string
	CWD         string
	Strict      bool
	Ptrace      bool
	EncodedArgs string
	Self        string
	Args        []string
}

const BuildCommandHelp = `
此命令用于构建或进入构建环境，配置并执行与运行时环境一致的构建过程。

# 使用默认选项进入构建环境shell
<program> build

# 进入构建环境，修正系统调用（chown）以便在容器内避免文件权限引起的失败。
# 注意：该选项将极大降低构建环境中程序的性能
<program> build --ptrace -- bash

# 使用 fuse-overlayfs 进行构建，指定 fuse-overlayfs 命令路径
<program> build --fuse-overlayfs /path/to/fuse-overlayfs

# 使用 fuse-overlayfs，并传入额外的命令参数
<program> build --fuse-overlayfs-args "--option=value"

# 启用严格模式，确保构建环境与运行时环境一致，且不包含开发工具（默认行为）
<program> build --strict

`

func GetBuildArgs() []string {
	args := []string{}

	if BuildFlag.RootFS != "" {
		args = append(args, "--rootfs", BuildFlag.RootFS)
	}

	if BuildFlag.CWD != "" {
		args = append(args, "--cwd", BuildFlag.CWD)
	}

	if BuildFlag.Strict {
		args = append(args, "--strict")
	}

	args = append(args, "--ptrace="+fmt.Sprint(BuildFlag.Ptrace))

	if GlobalFlag.FuseOverlayFS != "" {
		args = append(args, "--fuse-overlayfs", GlobalFlag.FuseOverlayFS)
	}

	if GlobalFlag.FuseOverlayFSArgs != "" {
		args = append(args, "--fuse-overlayfs-args", GlobalFlag.FuseOverlayFSArgs)
	}

	if BuildFlag.Self != "" {
		args = append(args, "--self", BuildFlag.Self)
	}

	if len(BuildFlag.Args) > 0 {
		args = append(args, "--")
		args = append(args, BuildFlag.Args...)
	}

	return args
}

func MountOverlayStage2() {

	overlayDir := path.Join(BuildFlag.CWD, FileSystemDir)
	mergedDir := path.Join(overlayDir, "merged")
	tmpRootFS := BuildFlag.TmpRootFS
	err := syscall.PivotRoot(tmpRootFS+mergedDir, tmpRootFS+mergedDir+BuildFlag.RootFS)
	if err != nil {
		log.Fatalln("PivotRoot2:", err)
	}
	if BuildFlag.Ptrace {
		Ptrace(BuildFlag.Self, BuildFlag.Args)
	} else {
		Exec(BuildFlag.Args...)
	}

}
func MountOverlay() {
	overlayDir := path.Join(BuildFlag.CWD, FileSystemDir)
	aptCacheDir := path.Join(BuildFlag.CWD, AptCacheDir)
	aptDataDir := path.Join(BuildFlag.CWD, AptDataDir)
	tmpRootFS := BuildFlag.TmpRootFS
	upperDir := path.Join(overlayDir, UpperDirName)
	lowerDir := path.Join(overlayDir, LowerDirName)
	workDir := path.Join(overlayDir, WorkDirName)
	mergedDir := path.Join(overlayDir, MergedDirName)
	cwdRootFSPivoted := fmt.Sprint(BuildFlag.RootFS, tmpRootFS)
	err := MkdirAlls([]string{tmpRootFS, upperDir, lowerDir, workDir, mergedDir}, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	err = MountBind(BuildFlag.RootFS, BuildFlag.RootFS, 0)
	if err != nil {
		log.Fatalln(err)
	}

	err = MountBind("/", tmpRootFS, 0)
	if err != nil {
		log.Fatalln(err)
	}

	err = MountAll([]MountOption{
		{
			Source: "sources.list.d",
			Target: lowerDir + "/etc/apt/sources.list.d",
		},
		{
			Source: "sources.list",
			Target: lowerDir + "/etc/apt/sources.list",
		},
		{
			Source: "apt.conf",
			Target: lowerDir + "/etc/apt/apt.conf",
		},
		{
			Source: "apt.conf.d",
			Target: lowerDir + "/etc/apt/apt.conf.d",
		},
	})
	if err != nil {
		log.Fatalln("MountAll:", err)
	}
	err = syscall.PivotRoot(BuildFlag.RootFS, cwdRootFSPivoted)
	if err != nil {
		log.Fatalln("PivotRoot:", BuildFlag.RootFS, cwdRootFSPivoted, err)
	}
	fuseOverlayFSOption := fmt.Sprintf("lowerdir=%s:%s,upperdir=%s,workdir=%s,squash_to_root",
		tmpRootFS+lowerDir,
		tmpRootFS+tmpRootFS,
		tmpRootFS+upperDir,
		tmpRootFS+workDir)
	fuseOverlayFSArgs := []string{"-o", fuseOverlayFSOption, tmpRootFS + mergedDir}
	if GlobalFlag.FuseOverlayFSArgs != "" {
		fuseOverlayFSArgs = append(fuseOverlayFSArgs, strings.Split(GlobalFlag.FuseOverlayFSArgs, " ")...)
	}
	err = RunCommand(GlobalFlag.FuseOverlayFS, fuseOverlayFSArgs...)
	if err != nil {
		log.Fatalln("fuse-overlayfs:", fuseOverlayFSOption, tmpRootFS+mergedDir, err)
	}
	err = MountAll([]MountOption{
		{
			Source: tmpRootFS + "/dev",
			Target: path.Join(tmpRootFS+mergedDir, "dev"),
		},
		{
			Source: tmpRootFS + "/proc",
			Target: path.Join(tmpRootFS+mergedDir, "proc"),
		},
		{
			Source: tmpRootFS + "/home",
			Target: path.Join(tmpRootFS+mergedDir, "home"),
		},
		{
			Source: tmpRootFS + "/project",
			Target: path.Join(tmpRootFS+mergedDir, "project"),
		},
		{
			Source: tmpRootFS + "/root",
			Target: path.Join(tmpRootFS+mergedDir, "root"),
		},
		{
			Source: tmpRootFS + "/tmp",
			Target: path.Join(tmpRootFS+mergedDir, "tmp"),
		},
		{
			Source: tmpRootFS + aptDataDir,
			Target: path.Join(tmpRootFS+mergedDir, "/var/lib/apt"),
		},
		{
			Source: tmpRootFS + aptCacheDir,
			Target: path.Join(tmpRootFS+mergedDir, "/var/cache"),
		},
	})
	if err != nil {
		log.Fatalln("MountAll2:", err)
	}
	SwitchTo("MountOverlayStage2", &SwitchFlags{Cloneflags: syscall.CLONE_NEWNS})
}

func BuildMain(cmd *cobra.Command, args []string) error {
	BuildFlag.Args = args
	reexec.Register("MountOverlay", MountOverlay)
	reexec.Register("MountOverlayStage2", MountOverlayStage2)
	if !reexec.Init() {
		if BuildFlag.Strict {
			encodedArgs := []string{}
			args := GetBuildArgs()
			for _, str := range args {
				encoded := base64.StdEncoding.EncodeToString([]byte(str))
				encodedArgs = append(encodedArgs, encoded)
			}
			extArgs := []string{"ll-builder", "run", "--exec", fmt.Sprintf("%s build --encoded-args %s", BuildFlag.Self, strings.Join(encodedArgs, ","))}
			Exec(extArgs...)
		}
		if BuildFlag.EncodedArgs != "" {
			args := []string{}
			for _, item := range strings.Split(BuildFlag.EncodedArgs, ",") {
				value, err := base64.StdEncoding.DecodeString(item)
				if err != nil {
					log.Fatalln(err)
				}
				args = append(args, string(value))
			}

			args = append([]string{"build"}, args...)
			return SwitchTo("MountOverlay", &SwitchFlags{
				UID:           0,
				GID:           0,
				Cloneflags:    syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
				Args:          args,
				NoDefaultArgs: true,
			})
		} else {
			args := GetBuildArgs()
			args = append([]string{"build"}, args...)
			return SwitchTo("MountOverlay", &SwitchFlags{
				UID:           0,
				GID:           0,
				Cloneflags:    syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
				Args:          args,
				NoDefaultArgs: true,
			})

		}
	}
	return nil
}

func CreateBuildCommand() *cobra.Command {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "构建或进入构建环境",
		Run: func(cmd *cobra.Command, args []string) {
			ExitWith(BuildMain(cmd, args))
		},
	}
	cmd.Flags().StringVar(&BuildFlag.RootFS, "rootfs", "/run/host/rootfs", "主机根目录路径")
	cmd.Flags().StringVar(&BuildFlag.TmpRootFS, "tmp-rootfs", "/tmp/rootfs", "临时根目录路径")
	cmd.Flags().StringVar(&BuildFlag.CWD, "cwd", cwd, "当前工作目录路径")
	if ptrace.IsSupported {
		cmd.Flags().BoolVar(&BuildFlag.Ptrace, "ptrace", false, "修正系统调用(chown)")
	}
	cmd.Flags().StringVar(&BuildFlag.EncodedArgs, "encoded-args", "", "编码后的参数")
	cmd.Flags().StringVar(&BuildFlag.Self, "self", os.Args[0], "ll-killer路径")
	cmd.Flags().BoolVarP(&BuildFlag.Strict, "strict", "x", os.Getenv("LINGLONG_APPID") == "", "严格模式，启动一个与运行时环境相同的构建环境，确保环境一致性（不含gcc等工具）")
	cmd.Flags().StringVar(&GlobalFlag.FuseOverlayFS, "fuse-overlayfs", "fuse-overlayfs", "fuse-overlayfs命令路径")
	cmd.Flags().StringVar(&GlobalFlag.FuseOverlayFSArgs, "fuse-overlayfs-args", "", "fuse-overlayfs命令额外参数")

	cmd.Flags().MarkHidden("encoded-args")
	cmd.Flags().MarkHidden("self")
	cmd.Flags().MarkHidden("cwd")
	return cmd
}
