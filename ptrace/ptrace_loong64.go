//go:build linux && loong64

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
	if regs.Regs[11] == syscall.SYS_FCHOWN || regs.Regs[11] == syscall.SYS_FCHOWNAT {
		if os.Getuid() != int(regs.Regs[6]) || os.Getgid() != int(regs.Regs[7]) {
			regs.Regs[6] = uint64(os.Getuid())
			regs.Regs[7] = uint64(os.Getgid())
			err := unix.PtraceSetRegs(pid, &regs)
			if err != nil {
				return fmt.Errorf("PtraceSetRegs:%v", err)
			}
		}
	}
	return nil
}
