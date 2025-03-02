//go:build linux && arm64

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
	if regs.Regs[8] == syscall.SYS_FCHOWN {
		if os.Getuid() != int(regs.Regs[1]) || os.Getgid() != int(regs.Regs[2]) {
			regs.Regs[1] = ^uint64(0)
			regs.Regs[2] = ^uint64(0)
			err := unix.PtraceSetRegs(pid, &regs)
			if err != nil {
				return fmt.Errorf("PtraceSetRegs:%v", err)
			}
		}
	}
	if regs.Regs[8] == syscall.SYS_FCHOWNAT {
		if os.Getuid() != int(regs.Regs[2]) || os.Getgid() != int(regs.Regs[3]) {
			regs.Regs[2] = ^uint64(0)
			regs.Regs[3] = ^uint64(0)
			err := unix.PtraceSetRegs(pid, &regs)
			if err != nil {
				return fmt.Errorf("PtraceSetRegs:%v", err)
			}
		}
	}
	return nil
}
