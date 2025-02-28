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
	"os/exec"
	"runtime"
	"syscall"

	ptrace "ll-killer/ptrace"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

const PtraceCommandHelp = `
此命令用于拦截并修正系统调用，当前仅处理chown调用。

# 用法
<program> ptrace -- 要处理的命令
`

func Ptrace(self string, args []string) {
	args = append([]string{self, "ptrace"}, args...)
	Exec(args...)
}
func PtraceMain(cmd *cobra.Command, args []string) error {
	var usage unix.Rusage
	var wstatus unix.WaitStatus
	var wpid int
	runtime.LockOSThread()
	if len(args) == 0 {
		args = []string{DefaultShell()}
	}

	binary, err := exec.LookPath(args[0])
	if err == nil {
		args[0] = binary
	}

	process, err := os.StartProcess(args[0], args, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Env:   os.Environ(),
		Sys: &syscall.SysProcAttr{
			Ptrace: true,
		},
	})
	if err != nil {
		return err
	}

	pid := process.Pid
	wpid, err = unix.Wait4(pid, &wstatus, syscall.WUNTRACED, &usage)
	if err != nil {
		return err
	}
	err = unix.PtraceSetOptions(pid,
		syscall.PTRACE_O_TRACECLONE|
			syscall.PTRACE_O_TRACESYSGOOD|
			syscall.PTRACE_O_TRACEFORK|
			syscall.PTRACE_O_TRACEEXEC|
			syscall.PTRACE_O_TRACEVFORK)
	if err != nil {
		return err
	}

	err = unix.PtraceSyscall(pid, 0)
	if err != nil {
		return err
	}
	// chown 4:5 test
	for {
		// syscall.RawSyscall(syscall.SYS_WAITID)
		wpid, err = unix.Wait4(-1, &wstatus, 0, &usage)
		if err != nil {
			return fmt.Errorf("Wait4:%v", err)
		}
		if wstatus.Exited() {
			if wpid == pid {
				return nil
			}
			continue
		}
		if wstatus.Stopped() && wstatus.StopSignal()&0x80 == 0 {
			err = unix.PtraceSyscall(wpid, 0)
			if err != nil {
				return fmt.Errorf("PtraceSyscall.SIGTRAP:%v", err)
			}
			continue
		}

		var regs unix.PtraceRegs
		err = unix.PtraceGetRegs(wpid, &regs)
		if err != nil {
			return fmt.Errorf("PtraceGetRegs:%v", err)
		}
		err = ptrace.PtraceHandle(wpid, regs)
		if err != nil {
			return fmt.Errorf("PtraceHandle:%v", err)
		}
		err = unix.PtraceSyscall(wpid, 0)

		if err != nil {
			return fmt.Errorf("PtraceSyscall:%v", err)
		}
	}
}
func CreatePtraceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ptrace",
		Short: "修正系统调用(chown)",
		Long:  BuildHelpMessage(PtraceCommandHelp),
		Run: func(cmd *cobra.Command, args []string) {
			ExitWith(PtraceMain(cmd, args))
		},
	}

	return cmd
}
