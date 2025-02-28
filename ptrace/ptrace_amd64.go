//go:build linux && amd64

/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package ptrace

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

var IsSupported = true

func PtraceHandle(pid int, regs unix.PtraceRegs) error {
	if regs.Orig_rax == syscall.SYS_CHOWN || regs.Orig_rax == syscall.SYS_FCHOWN || regs.Orig_rax == syscall.SYS_FCHOWNAT {
		if os.Getuid() != int(regs.Rdx) || os.Getgid() != int(regs.R10) {
			regs.Rdx = uint64(os.Getuid())
			regs.R10 = uint64(os.Getgid())
			err := unix.PtraceSetRegs(pid, &regs)
			if err != nil {
				return fmt.Errorf("PtraceSetRegs:%v", err)
			}
		}
	}
	return nil
}
