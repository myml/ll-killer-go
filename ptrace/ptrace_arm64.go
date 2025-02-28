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
	if regs.Regs[8] == syscall.SYS_FCHOWN || regs.Regs[8] == syscall.SYS_FCHOWNAT {
		if os.Getuid() != int(regs.Regs[2]) || os.Getgid() != int(regs.Regs[3]) {
			regs.Regs[2] = uint64(os.Getuid())
			regs.Regs[3] = uint64(os.Getgid())
			err := unix.PtraceSetRegs(pid, &regs)
			if err != nil {
				return fmt.Errorf("PtraceSetRegs:%v", err)
			}
		}
	}
	return nil
}
