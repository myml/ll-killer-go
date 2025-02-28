//go:build linux && !amd64 && !arm64 && !loong64

/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */

package ptrace

import (
	"golang.org/x/sys/unix"
)

var IsSupported = false

func PtraceHandle(pid int, regs unix.PtraceRegs) error {
	return nil
}
