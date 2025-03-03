/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/moby/sys/reexec"
)

const (
	KillerExec        = "ll-killer"
	KillerExecEnv     = "KILLER_EXEC"
	FileSystemDir     = "linglong/filesystem"
	UpperDirName      = "diff"
	LowerDirName      = "overwrite"
	WorkDirName       = "work"
	MergedDirName     = "merged"
	SourceListFile    = "sources.list"
	AptDir            = "linglong/apt"
	AptDataDir        = AptDir + "/data"
	AptCacheDir       = AptDir + "/cache"
	AptConfDir        = "apt.conf.d"
	AptConfFile       = AptConfDir + "/ll-killer.conf"
	kLinglongYaml     = "linglong.yaml"
	kKillerCommands   = "KILLER_COMMANDS"
	kKillerDebug      = "KILLER_DEBUG"
	kMountArgsSep     = ":"
	kMountArgsItemSep = "+"
	FuseOverlayFSType = "fuse-overlayfs"
)

var (
	Version   = "unknown"
	BuildTime = "unknown"
)

var GlobalFlag = struct {
	Debug             bool
	FuseOverlayFS     string
	FuseOverlayFSArgs string
}{
	Debug:             false,
	FuseOverlayFS:     "fuse-overlayfs",
	FuseOverlayFSArgs: "",
}

type SwitchFlags struct {
	UID           int
	GID           int
	Cloneflags    uintptr
	Args          []string
	NoDefaultArgs bool
	UidMappings   []string
	GidMappings   []string
}

func CreateCommand(name string) *exec.Cmd {
	cmd := reexec.Command(name)
	cmd.Args = append(cmd.Args, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

/*
1000->0->1000 500:500:1000

1000->0
1000:0:1
500:100000+500:500
1001:100000+1001:499

0->1000
0:1000:1
100000+500:500:500
100000+1001:1001:499

另一个映射版本：
1000:1000:1
1001:100000:65535

1000->0
0:1000:1
1:100000:


*/

// mergeMappings 合并映射表中相邻连续的映射项
// func mergeMappings(mappings []syscall.SysProcIDMap) []syscall.SysProcIDMap {
// 	sort.Slice(mappings, func(i, j int) bool {
// 		return mappings[i].HostID < mappings[j].HostID
// 	})
// 	if len(mappings) == 0 {
// 		return mappings
// 	}

// 	merged := []syscall.SysProcIDMap{mappings[0]}
// 	for i := 1; i < len(mappings); i++ {
// 		last := &merged[len(merged)-1]
// 		current := mappings[i]
// 		if last.HostID+last.Size == current.HostID &&
// 			last.ContainerID+last.Size == current.ContainerID {
// 			last.Size += current.Size
// 		} else {
// 			merged = append(merged, current)
// 		}
// 	}
// 	return merged
// }
// func SplitMapping(from int, to int, mappings []syscall.SysProcIDMap) []syscall.SysProcIDMap {
// 	var newMappings []syscall.SysProcIDMap
// 	for _, item := range mappings {
// 		if item.HostID <= from && from < item.HostID+item.Size {
// 			if item.HostID == from && item.ContainerID == to {
// 				newMappings = append(newMappings, item)
// 			} else {
// 				newMappings = append(newMappings, syscall.SysProcIDMap{
// 					HostID:      from,
// 					ContainerID: to,
// 					Size:        1,
// 				})
// 				if item.HostID < from {
// 					newMappings = append(newMappings, syscall.SysProcIDMap{
// 						HostID:      item.HostID,
// 						ContainerID: item.ContainerID,
// 						Size:        from - item.HostID,
// 					})
// 				}
// 				if item.HostID+item.Size-1 > from {
// 					diff := from - item.HostID
// 					newMappings = append(newMappings, syscall.SysProcIDMap{
// 						HostID:      item.HostID + diff + 1,
// 						ContainerID: item.ContainerID + diff + 1,
// 						Size:        item.HostID + item.Size - from - 1,
// 					})
// 				}
// 			}
// 		} else {
// 			newMappings = append(newMappings, item)
// 		}
// 	}

// 	return mergeMappings(newMappings)

// }
func SwitchTo(next string, flags *SwitchFlags) error {
	cmd := CreateCommand(next)
	if flags.NoDefaultArgs {
		cmd.Args = []string{cmd.Args[0]}
	}
	if len(flags.Args) > 0 {
		cmd.Args = append(cmd.Args, flags.Args...)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Unshareflags: flags.Cloneflags,
	}
	if flags.Cloneflags&syscall.CLONE_NEWUSER != 0 {
		if os.Getuid() != flags.UID && os.Getgid() != flags.GID {
			cmd.SysProcAttr.UidMappings = []syscall.SysProcIDMap{
				{
					ContainerID: flags.UID,
					HostID:      os.Getuid(),
					Size:        1,
				},
			}
			cmd.SysProcAttr.GidMappings = []syscall.SysProcIDMap{
				{
					ContainerID: flags.GID,
					HostID:      os.Getgid(),
					Size:        1,
				},
			}
		} else {
			cmd.SysProcAttr.Unshareflags ^= syscall.CLONE_NEWUSER
		}
	}
	Debug("SwitchTo", fmt.Sprintf("%#x", cmd.SysProcAttr.Unshareflags), cmd.Path, cmd.Args)
	return cmd.Run()
}

func RunCommand(name string, args ...string) error {
	Debug("RunCommand", name, args)
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func DefaultShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "/bin/bash"
	}
	return shell
}
func ExecRaw(args ...string) error {
	Debug("Exec", args)
	if len(args) == 0 {
		args = []string{DefaultShell()}
	}

	binary, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("%s:%s", binary, err)
	}

	if err := syscall.Exec(binary, args, os.Environ()); err != nil {
		return fmt.Errorf("%s:%s", binary, err)
	}
	return nil
}
func Exec(args ...string) {
	err := ExecRaw(args...)
	if err != nil {
		log.Fatalln(err)
	}
}
func IsExist(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func Mount(opt *MountOption) error {
	if opt.FSType == "" && (opt.Flags == 0 || opt.Flags&syscall.MS_BIND != 0) {
		return MountBind(opt.Source, opt.Target, opt.Flags)
	}
	return syscall.Mount(opt.Source, opt.Target, opt.FSType, uintptr(opt.Flags), opt.Data)
}
func MountBind(source string, target string, flags int) error {
	Debug("MountBind", source, target, flags)
	srcInfo, err := os.Lstat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(source)
		if err != nil {
			return err
		}
		dirname := path.Dir(target)
		err = os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
		os.Remove(target)
		err = os.Symlink(link, target)
		if err == nil {
			return nil
		}
		if !os.IsExist(err) {
			return err
		}
	}

	if !IsExist(target) {
		if srcInfo.IsDir() {
			err = os.MkdirAll(target, 0755)
			if err != nil {
				return err
			}
		} else {
			dirname := path.Dir(target)
			err = os.MkdirAll(dirname, 0755)
			if err != nil {
				return err
			}

			_, err = os.Create(target)
			if err != nil {
				return err
			}
		}
	}
	flags |= syscall.MS_BIND | syscall.MS_PRIVATE
	if srcInfo.IsDir() {
		flags |= syscall.MS_REC
	}
	err = syscall.Mount(source, target, "none", uintptr(flags), "")
	if err != nil {
		return fmt.Errorf("mount:%s", err)
	}
	return nil
}
func MkdirAlls(dirs []string, mode os.FileMode) error {
	for _, dir := range dirs {
		err := os.MkdirAll(dir, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

type MountOption struct {
	Source string
	Target string
	FSType string
	Flags  int
	Data   string
}

func MountAll(opts []MountOption) error {
	for _, opt := range opts {
		if err := opt.Mount(); err != nil {
			return err
		}
	}
	return nil
}

var MountFlagMap = map[string]int{
	"active":      syscall.MS_ACTIVE,
	"async":       syscall.MS_ASYNC,
	"bind":        syscall.MS_BIND,
	"rbind":       syscall.MS_BIND | syscall.MS_REC,
	"dirsync":     syscall.MS_DIRSYNC,
	"invalidate":  syscall.MS_INVALIDATE,
	"i_version":   syscall.MS_I_VERSION,
	"kernmount":   syscall.MS_KERNMOUNT,
	"mandlock":    syscall.MS_MANDLOCK,
	"move":        syscall.MS_MOVE,
	"noatime":     syscall.MS_NOATIME,
	"nodev":       syscall.MS_NODEV,
	"nodiratime":  syscall.MS_NODIRATIME,
	"noexec":      syscall.MS_NOEXEC,
	"nosuid":      syscall.MS_NOSUID,
	"nouser":      syscall.MS_NOUSER,
	"posixacl":    syscall.MS_POSIXACL,
	"private":     syscall.MS_PRIVATE,
	"rdonly":      syscall.MS_RDONLY,
	"rec":         syscall.MS_REC,
	"relatime":    syscall.MS_RELATIME,
	"remount":     syscall.MS_REMOUNT,
	"shared":      syscall.MS_SHARED,
	"silent":      syscall.MS_SILENT,
	"slave":       syscall.MS_SLAVE,
	"strictatime": syscall.MS_STRICTATIME,
	"sync":        syscall.MS_SYNC,
	"synchronous": syscall.MS_SYNCHRONOUS,
	"unbindable":  syscall.MS_UNBINDABLE,
}

func ParseMountFlag(flag string) int {
	flags := strings.Split(flag, kMountArgsItemSep)
	value := 0
	for _, flag := range flags {
		if flag == "" {
			continue
		}
		item, ok := MountFlagMap[flag]
		if !ok {
			log.Println("ignore mount flag:", flag)
			continue
		}
		value |= item
	}
	return value
}
func ParseMountOption(item string) MountOption {
	chunks := strings.SplitN(item, kMountArgsSep, 5)
	opt := MountOption{}
	if len(chunks) >= 2 {
		opt.Source = chunks[0]
		opt.Target = chunks[1]
	}
	if len(chunks) >= 3 {
		opt.Flags = ParseMountFlag(chunks[2])
	}
	if len(chunks) >= 4 {
		opt.FSType = chunks[3]
	}
	if len(chunks) >= 5 {
		opt.Data = chunks[4]
	}
	return opt
}

func (opt *MountOption) buildFileSystem(sources []string, target string, filesystem map[string]string, excludes []string) (map[string]string, error) {
	conflicts := make(map[string]struct{})
	for _, source := range sources {
		for _, prefix := range excludes {
			if source == prefix {
				Debug("mount.skip", source, target)
				filesystem[target] = source
				return filesystem, nil
			}
		}
	}

	items := make([][]os.DirEntry, len(sources))
	for i, dir := range sources {
		files, _ := os.ReadDir(dir)
		items[i] = files
	}

	for i, parent := range sources {
		files := items[i]
		if files == nil {
			continue
		}

		for _, file := range files {
			path := filepath.Join(target, file.Name())
			current := filepath.Join(parent, file.Name())
			if _, exists := conflicts[file.Name()]; exists {
				continue
			}
			if _, exists := filesystem[path]; exists && file.IsDir() {
				conflicts[file.Name()] = struct{}{}
				delete(filesystem, path)
				continue
			}
			filesystem[path] = current
		}
	}

	for name := range conflicts {
		var walks []string
		for i, dir := range sources {
			if items[i] != nil {
				next := filepath.Join(dir, name)
				walks = append(walks, next)
			}
		}
		_, err := opt.buildFileSystem(walks, filepath.Join(target, name), filesystem, excludes)
		if err != nil {
			return nil, err
		}
	}
	return filesystem, nil
}

func (opt *MountOption) Mount() error {
	Debug("Mount", []string{opt.Source, opt.Target, opt.FSType, opt.FSType, opt.Data})
	if opt.FSType == "merge" {
		filesystem := make(map[string]string)
		excludes := append([]string{opt.Target}, strings.Split(opt.Data, kMountArgsItemSep)...)
		if opt.Data == "" {
			excludes = append(excludes, "/tmp", "/proc", "/dev", "/sys", "/run", "/var/run")
		}
		for index, path := range excludes {
			result, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			excludes[index] = result
		}
		target, err := filepath.Abs(opt.Target)
		if err != nil {
			return err
		}
		sources := strings.Split(opt.Source, kMountArgsItemSep)
		for index, path := range sources {
			result, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			sources[index] = result
		}
		filesystem, err = opt.buildFileSystem(sources, target, filesystem, excludes)
		if err != nil {
			return err
		}

		for mount, from := range filesystem {
			if err = MountBind(from, mount, opt.Flags); err != nil {
				return err
			}
		}
		return nil
	}
	if opt.FSType == FuseOverlayFSType {
		fuseOverlayFSArgs := []string{"-o", opt.Data, opt.Target}
		if GlobalFlag.FuseOverlayFSArgs != "" {
			fuseOverlayFSArgs = append(fuseOverlayFSArgs, strings.Split(GlobalFlag.FuseOverlayFSArgs, " ")...)
		}
		dirs := []string{opt.Target}
		for _, item := range strings.Split(opt.Data, ",") {
			param := strings.SplitN(item, "=", 2)
			if len(param) == 2 {
				key := strings.TrimSpace(param[0])
				value := strings.TrimSpace(param[1])
				if key == "upperdir" || key == "workdir" {
					dirs = append(dirs, value)
				}
			}
		}
		Debug("mkdir.overlay", dirs)
		if err := MkdirAlls(dirs, 0755); err != nil {
			log.Println(err)
		}
		return RunCommand(GlobalFlag.FuseOverlayFS, fuseOverlayFSArgs...)
	}
	return Mount(opt)
}

func Debug(v ...any) {
	if GlobalFlag.Debug {
		log.Println(v...)
	}
}
func ExitWith(err error) {
	if err == nil {
		os.Exit(0)
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	log.Fatalln(err)
}

func WriteFile(name string, data []byte, perm os.FileMode) error {
	if perm&os.FileMode(os.O_EXCL) != 0 && IsExist(name) {
		return nil
	}
	return os.WriteFile(name, data, perm)
}
func BuildHelpMessage(help string) string {
	return strings.ReplaceAll(help, "<program>", os.Args[0])
}
func SetupEnvVar() error {
	if os.Getenv(KillerExecEnv) == "" {
		path, err := exec.LookPath(KillerExec)
		if err != nil {
			path, err = filepath.Abs(os.Args[0])
			if err != nil {
				return err
			}
			os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), filepath.Dir(path)))
		}
		os.Setenv(KillerExecEnv, path)
	}
	return nil
}
func DumpYaml(file string, v interface{}) error {
	fs, err := os.Create(file)
	if err != nil {
		return err
	}
	encoder := yaml.NewEncoder(fs)
	defer encoder.Close()
	err = encoder.Encode(v)
	if err != nil {
		return err
	}
	return nil
}

func CopyFileIO(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
