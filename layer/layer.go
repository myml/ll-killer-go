package layer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"ll-killer/utils"
	"os"
	"path"
	"strings"
)

const LayerMagic = "<<< deepin linglong layer archive >>>\x00\x00\x00"
const MkfsErofs = "mkfs.erofs"
const DumpErofs = "dump.erofs"
const ErofsFuse = "erofsfuse"
const FuserMount = "fusermount"

type LayerInfo struct {
	Arch          []string `json:"arch"`
	Base          string   `json:"base"`
	Channel       string   `json:"channel"`
	Description   string   `json:"description"`
	ID            string   `json:"id"`
	Kind          string   `json:"kind"`
	Module        string   `json:"module"`
	Name          string   `json:"name"`
	SchemaVersion string   `json:"schema_version"`
	Size          int      `json:"size"`
	Version       string   `json:"version"`
}
type LayerInfoHeader struct {
	Info    LayerInfo `json:"info"`
	Version string    `json:"version"`
}

func (l *LayerInfoHeader) Print() {
	fmt.Println("Layer元数据版本:")
	fmt.Printf("    版本: %s\n", l.Version)
	l.Info.Print()
}

func (l *LayerInfo) FileName() string {
	return fmt.Sprintf("%s_%s_%s_%s.layer", l.ID, l.Version, strings.Join(l.Arch, "+"), l.Module)
}

func (info *LayerInfo) Print() {
	fmt.Println("Layer元数据:")
	fmt.Printf("    名称: %s\n", info.Name)
	fmt.Printf("    ID: %s\n", info.ID)
	fmt.Printf("    版本: %s\n", info.Version)
	fmt.Printf("    模块: %s\n", info.Module)
	fmt.Printf("    类型: %s\n", info.Kind)
	fmt.Printf("    基础: %s\n", info.Base)
	fmt.Printf("    渠道: %s\n", info.Channel)
	fmt.Printf("    元数据版本: %s\n", info.SchemaVersion)
	fmt.Printf("    描述: %s\n", info.Description)
	fmt.Printf("    大小: %d 字节\n", info.Size)
	fmt.Printf("    架构: %s\n", strings.Join(info.Arch, ", "))
}

type LayerHeader struct {
	FileName string
	FileSize int64
	Magic    string
	Info     LayerInfoHeader
	InfoSize int
}

func (l *LayerHeader) DataOffset() int {
	return 40 + 4 + l.InfoSize
}
func (l *LayerHeader) PrintAll() error {
	l.Print()
	l.Info.Print()
	return l.PrintErofs(nil)
}
func (l *LayerHeader) Print() {
	fmt.Println("Layer文件头:")
	fmt.Printf("  文件名: %s\n", l.FileName)
	fmt.Printf("  文件大小: %d 字节\n", l.FileSize)
	fmt.Printf("  魔数: %s\n", l.Magic)
	fmt.Printf("  元数据大小: %d 字节\n", l.InfoSize)
	fmt.Printf("  数据偏移量: %d\n", l.DataOffset())
}

type DumpErofsOption struct {
	Args     []string
	ExecPath string
}

func (l *LayerHeader) PrintErofs(opt *DumpErofsOption) error {
	execPath := DumpErofs
	args := []string{fmt.Sprint("--offset=", l.DataOffset()), l.FileName}
	if opt != nil {
		if opt.ExecPath != "" {
			execPath = opt.ExecPath
		}
		args = append(args, opt.Args...)
	}
	fmt.Println("Erofs信息:")
	return utils.RunCommand(execPath, args...)
}

func NewLayerHeader(file *os.File) (*LayerHeader, error) {
	var layer LayerHeader
	file.Seek(0, io.SeekStart)
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	layer.FileSize = info.Size()
	layer.FileName = info.Name()

	magic := make([]byte, 40)
	_, err = io.ReadFull(file, magic)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(magic, []byte(LayerMagic)) {
		return nil, fmt.Errorf("错误的文件头：%v %v", magic, []byte(LayerMagic))
	}
	layer.Magic = string(magic)

	var metadataSize int32
	err = binary.Read(file, binary.LittleEndian, &metadataSize)
	if err != nil {
		return nil, err
	}
	layer.InfoSize = int(metadataSize)

	metadataBytes := make([]byte, metadataSize)
	_, err = io.ReadFull(file, metadataBytes)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metadataBytes, &layer.Info)
	if err != nil {
		return nil, err
	}

	return &layer, nil
}
func NewLayerHeaderFromFile(filePath string) (*LayerHeader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	layer, err := NewLayerHeader(file)
	layer.FileName = filePath
	return layer, err
}

type MountOption struct {
	Source   string
	Target   string
	Args     []string
	ExecPath string
}

func Mount(opt *MountOption) error {
	execPath := ErofsFuse
	if opt.ExecPath != "" {
		execPath = opt.ExecPath
	}
	header, err := NewLayerHeaderFromFile(opt.Source)
	if err != nil {
		return err
	}
	err = os.MkdirAll(opt.Target, 0755)
	if err != nil {
		return err
	}
	args := []string{fmt.Sprint("--offset=", header.DataOffset()), opt.Source, opt.Target}
	if len(opt.Args) > 0 {
		args = append(args, opt.Args...)
	}
	return utils.RunCommand(execPath, args...)
}

type UmountOption struct {
	Target   string
	Args     []string
	ExecPath string
}

func Umount(opt *UmountOption) error {
	execPath := FuserMount
	if opt.ExecPath != "" {
		execPath = opt.ExecPath
	}
	args := []string{"-u", opt.Target}
	if len(opt.Args) > 0 {
		args = append(args, opt.Args...)
	}
	return utils.RunCommand(execPath, args...)
}

type PackOption struct {
	Source     string
	Target     string
	ExecPath   string
	Compressor string
	BlockSize  int
	Args       []string
}

func Pack(opt *PackOption) error {
	execPath := MkfsErofs
	var layer LayerHeader
	layer.Info.Version = "1"
	infoJson, err := os.Open(path.Join(opt.Source, "info.json"))
	if err != nil {
		return err
	}
	infoData, err := io.ReadAll(infoJson)
	defer infoJson.Close()
	err = json.Unmarshal(infoData, &layer.Info.Info)
	if err != nil {
		return err
	}
	target := opt.Target
	if target == "" {
		target = layer.Info.Info.FileName()
	}
	if opt.ExecPath != "" {
		execPath = opt.ExecPath
	}
	metadata, err := json.Marshal(layer.Info)
	if err != nil {
		return err
	}
	layer.InfoSize = len(metadata)
	args := []string{}
	args = append(args, opt.Args...)
	if opt.Compressor != "" {
		args = append(args, fmt.Sprint("-z", opt.Compressor))
	}

	sizeData := make([]byte, 4)
	_, err = binary.Encode(sizeData, binary.LittleEndian, int32(layer.InfoSize))
	if err != nil {
		return err
	}

	args = append(args,
		fmt.Sprint("--offset=", layer.DataOffset()),
		fmt.Sprint("-b", opt.BlockSize),
		target,
		opt.Source)
	err = utils.RunCommand(execPath, args...)
	if err != nil {
		return err
	}

	header := append([]byte(LayerMagic), append(sizeData, metadata...)...)
	fp, err := os.OpenFile(target, os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.WriteAt(header, 0)
	if err != nil {
		return err
	}
	fmt.Println("文件已输出至:", target)
	return err
}

func Dump(target string) error {
	header, err := NewLayerHeaderFromFile(target)
	if err != nil {
		return err
	}
	return header.PrintAll()
}
