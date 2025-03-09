/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/moby/sys/reexec"
	"github.com/spf13/cobra"
)

/*
#cgo pkg-config: fuse3
#cgo LDFLAGS: -L. -lfuse-overlayfs -lgnu
#include <stdlib.h>
extern int fuse_ovl_main(int argc, char** argv);
*/
import "C"

var OverlayFlag struct {
	Self  string
	Shell string
	Args  []string
}

func FuseOvlMain(args []string) error {
	Debug("FuseOvlMain", args)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	argc := len(args) + 1
	argv := make([]*C.char, argc+1)
	args = append([]string{os.Args[0]}, args...)
	for i, arg := range args {
		argv[i] = C.CString(arg)
		defer C.free(unsafe.Pointer(argv[i]))
	}
	exitCode := int(C.fuse_ovl_main(C.int(argc), (**C.char)(unsafe.Pointer(&argv[0]))))
	if exitCode != 0 {
		return fmt.Errorf("fuse_ovl returned %d", exitCode)
	}
	return nil
}

func ExecFuseOvlMain(args []string) error {
	Debug("ExecFuseOvlMain", args)
	return RunCommand(reexec.Self(), append([]string{"overlay"}, args...)...)
}

type FuseOvlMountFlag struct {
	RedirectDir      string   // redirect_dir=%s
	Context          string   // context=%s
	LowerDir         []string // lowerdir=%s
	UpperDir         string   // upperdir=%s
	WorkDir          string   // workdir=%s
	UIDMapping       string   // uidmapping=%s
	GIDMapping       string   // gidmapping=%s
	Timeout          string   // timeout=%s
	Threaded         string   // threaded=%d,
	Fsync            string   // fsync=%d,
	FastIno          string   // fast_ino=%d
	Writeback        string   // writeback=%d,
	NoXattrs         string   // noxattrs=%d,出
	Plugins          string   // plugins=%s
	XattrPermissions string   // xattr_permissions=%d
	SquashToRoot     bool     // squash_to_root
	SquashToUID      string   // squash_to_uid=%d,
	SquashToGID      string   // squash_to_gid=%d,
	StaticNlink      bool     // static_nlink
	Volatile         bool     // volatile
	NoACL            bool     // noacl

	Target string   // 挂载目标
	Flags  []string // 额外的标志选项
	Args   []string // 附加参数
}

func FuseOvlMount(opt FuseOvlMountFlag) error {
	args := []string{}
	options := []string{}

	if opt.RedirectDir != "" {
		options = append(options, fmt.Sprint("redirect_dir=", opt.RedirectDir))
	}
	if opt.Context != "" {
		options = append(options, fmt.Sprint("context=", opt.Context))
	}
	if len(opt.LowerDir) > 0 {
		options = append(options, fmt.Sprint("lowerdir=", strings.Join(opt.LowerDir, ":")))
	}
	if opt.UpperDir != "" {
		options = append(options, fmt.Sprint("upperdir=", opt.UpperDir))
	}
	if opt.WorkDir != "" {
		options = append(options, fmt.Sprint("workdir=", opt.WorkDir))
	}
	if opt.UIDMapping != "" {
		options = append(options, fmt.Sprint("uidmapping=", opt.UIDMapping))
	}
	if opt.GIDMapping != "" {
		options = append(options, fmt.Sprint("gidmapping=", opt.GIDMapping))
	}
	if opt.Timeout != "" {
		options = append(options, fmt.Sprint("timeout=", opt.Timeout))
	}
	if opt.Threaded != "" {
		options = append(options, fmt.Sprint("threaded=", opt.Threaded))
	}
	if opt.Fsync != "" {
		options = append(options, fmt.Sprint("fsync=", opt.Fsync))
	}
	if opt.FastIno != "" {
		options = append(options, fmt.Sprint("fast_ino=", opt.FastIno))
	}
	if opt.Writeback != "" {
		options = append(options, fmt.Sprint("writeback=", opt.Writeback))
	}
	if opt.NoXattrs != "" {
		options = append(options, fmt.Sprint("noxattrs=", opt.NoXattrs))
	}
	if opt.Plugins != "" {
		options = append(options, fmt.Sprint("plugins=", opt.Plugins))
	}
	if opt.XattrPermissions != "" {
		options = append(options, fmt.Sprint("xattr_permissions=", opt.XattrPermissions))
	}
	if opt.SquashToRoot {
		options = append(options, "squash_to_root")
	}
	if opt.SquashToUID == "" {
		options = append(options, fmt.Sprint("squash_to_uid=", opt.SquashToUID))
	}
	if opt.SquashToGID == "" {
		options = append(options, fmt.Sprint("squash_to_gid=", opt.SquashToGID))
	}
	if opt.StaticNlink {
		options = append(options, "static_nlink")
	}
	if opt.Volatile {
		options = append(options, "volatile")
	}
	if opt.NoACL {
		options = append(options, "noacl")
	}

	options = append(options, opt.Flags...)

	if len(options) > 0 {
		args = append(args, "-o", strings.Join(options, ","))
	}
	args = append(args, opt.Target)
	args = append(args, opt.Args...)

	return FuseOvlMain(args)
}

func OverlayMain(cmd *cobra.Command, args []string) error {
	return FuseOvlMain(args)
}

func CreateOverlayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "overlay",
		Short:              "内置fuse-overlayfs挂载",
		Long:               "此命令调用内嵌的fuse-overlayfs程序实现overlay挂载",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			ExitWith(OverlayMain(cmd, args))
		},
	}

	return cmd
}
