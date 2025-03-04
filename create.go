/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type Config struct {
	Version string `yaml:"version"`
	Package struct {
		ID          string `yaml:"id"`
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Kind        string `yaml:"kind"`
		Description string `yaml:"description"`
	} `yaml:"package"`
	Command []string `yaml:"command"`
	Base    string   `yaml:"base"`
	Runtime string   `yaml:"runtime,omitempty"`
	Build   string   `yaml:"build"`
}

var ConfigData Config
var CreateFlag struct {
	NoBuild  bool
	Force    bool
	Metadata string
	Extend   string
}

const CreateCommandDescription = `创建一个玲珑应用项目，包括：
- 生成 linglong.yaml 配置文件
- 创建 build-aux 目录并填充辅助构建脚本
- 生成 apt.conf.d 和 sources.list.d 以支持包管理
- 自动执行一次 build 以初始化构建环境

注意：
- 如果 linglong.yaml 是手动创建的，需要手动运行 ll-builder build 进行初始化。
- 在未初始化的情况下，build 命令中的严格模式不可用。
- 关于辅助脚本的内容，请查看build-aux子命令帮助。

`
const CreateCommandHelp = `
# 在当前目录创建一个id为appId的项目
<program> create --id appId

# 从apt show的信息中提取字段，并创建项目
apt show app >pkg.info
<program> create --id appId --from pkg.info
`

// TODO: 单元测试
func NormalizeVersion(version string) string {
	re := regexp.MustCompile(`\D`)
	chunks := strings.SplitN(version, ".", 4)
	for index, chunk := range chunks {
		version := strings.TrimLeft(strings.TrimSpace(re.ReplaceAllString(chunk, "")), "0")
		if version == "" {
			version = "0"
		}
		chunks[index] = version
	}
	for len(chunks) < 4 {
		chunks = append(chunks, "0")
	}
	return strings.Join(chunks, ".")
}
func SetupPackageMetadata(cmd *cobra.Command) error {
	metadata, err := ParsePackageMetadataFromFile(CreateFlag.Metadata)
	if err != nil {
		return err
	}

	if !cmd.Flags().Changed("description") && metadata["description"] != "" {
		ConfigData.Package.Description = metadata["description"]
	}
	if !cmd.Flags().Changed("version") && metadata["version"] != "" {
		ConfigData.Package.Version = NormalizeVersion(metadata["version"])
	}
	if !cmd.Flags().Changed("id") && metadata["package"] != "" {
		ConfigData.Package.ID = metadata["package"]
	}
	if !cmd.Flags().Changed("name") && metadata["package"] != "" {
		ConfigData.Package.Name = metadata["package"]
	}
	if !cmd.Flags().Changed("base") && metadata["base"] != "" {
		ConfigData.Package.Name = metadata["runtime"]
	}
	if !cmd.Flags().Changed("runtime") && metadata["runtime"] != "" {
		ConfigData.Package.Name = metadata["runtime"]
	}
	if metadata["apt-sources"] != "" {
		if CreateFlag.Force || !IsExist(SourceListFile) {
			re := regexp.MustCompile(`^(http\S+?)\s+(\S+?)/(\S+)`)
			entries := strings.Split(metadata["apt-sources"], "\n")
			parsed := []string{}
			for _, entry := range entries {
				entry = strings.TrimSpace(entry)
				if !strings.HasPrefix(entry, "deb") {
					matched := re.FindStringSubmatch(entry)
					if len(matched) != 4 {
						log.Println("无效APT源:", entry)
						continue
					}
					url := fmt.Sprintf("%s/dists/%s/Release", matched[1], matched[2])
					release, err := ParsePackageMetadataFromUrl(url)
					if err != nil {
						log.Println(err)
					}
					if err == nil && release["components"] != "" {
						entry = fmt.Sprintf("deb [trusted=yes] %s %s %s", matched[1], matched[2], release["components"])
					} else {
						entry = fmt.Sprintf("deb [trusted=yes] %s %s %s", matched[1], matched[2], matched[3])
					}
				}
				parsed = append(parsed, entry)
			}
			if len(parsed) > 0 {
				err := WriteFile(SourceListFile, []byte(strings.Join(parsed, "\n")), 0755, CreateFlag.Force)
				if err != nil {
					return err
				}
				log.Println("created: ", SourceListFile)
			}
		} else {
			log.Println("skip: ", SourceListFile)
		}
	}
	return nil
}
func ParsePackageMetadataFromFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ParsePackageMetadata(file)
}
func ParsePackageMetadataFromUrl(url string) (map[string]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GET %s:%s", url, resp.Status)
	}
	defer resp.Body.Close()
	return ParsePackageMetadata(resp.Body)
}
func ParsePackageMetadata(stream io.Reader) (map[string]string, error) {
	metadata := make(map[string]string)

	scanner := bufio.NewScanner(stream)
	var key string
	var value string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if key != "" && strings.HasPrefix(line, " ") {
			line = strings.TrimSpace(line)
			if line == "." {
				line = ""
			}
			metadata[key] += "\n" + line
		} else {
			chunks := strings.SplitN(line, ":", 2)
			if len(chunks) < 2 {
				continue
			}

			key = strings.ToLower(strings.TrimSpace(chunks[0]))
			value = strings.TrimSpace(chunks[1])
			metadata[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return metadata, nil
}

func SetupProject(target string) error {
	ConfigData.Command[0] = strings.ReplaceAll(ConfigData.Command[0], "<APPID>", ConfigData.Package.ID)

	err := DumpYaml(kLinglongYaml, ConfigData)
	if err != nil {
		return err
	}
	log.Println("created:", kLinglongYaml)
	return nil
}

func CreateMain(cmd *cobra.Command, args []string) error {

	if err := ExtractBuildAuxFiles(CreateFlag.Force); err != nil {
		return err
	}

	if err := SetupPackageMetadata(cmd); err != nil {
		return err
	}

	if err := SetupProject(kLinglongYaml); err != nil {
		return err
	}

	if !CreateFlag.NoBuild {
		return RunCommand("ll-builder", "build", "--exec", "true")
	}

	return nil
}
func CreateCreateCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "创建模板项目",
		Example: BuildHelpMessage(CreateCommandHelp),
		Long:    BuildHelpMessage(CreateCommandDescription),
		Run: func(cmd *cobra.Command, args []string) {
			ExitWith(CreateMain(cmd, args))
		},
	}
	cmd.Flags().StringVar(&ConfigData.Version, "spec", "1", "玲珑yaml版本")
	cmd.Flags().StringVar(&ConfigData.Package.ID, "id", "app", "包名")
	cmd.Flags().StringVar(&ConfigData.Package.Name, "name", "app", "显示名称")
	cmd.Flags().StringVar(&ConfigData.Package.Version, "version", "0.0.0.1", "版本号")
	cmd.Flags().StringVar(&ConfigData.Package.Kind, "kind", "app", "应用类型：app|runtime")
	cmd.Flags().StringVar(&ConfigData.Package.Description, "description", "", "应用说明")
	cmd.Flags().StringArrayVar(&ConfigData.Command, "command", []string{"/opt/apps/<APPID>/files/entrypoint.sh"}, "启动命令")
	cmd.Flags().StringVar(&ConfigData.Base, "base", "org.deepin.base/23.1.0", "Base镜像")
	cmd.Flags().StringVar(&ConfigData.Runtime, "runtime", "", "Runtime镜像")
	cmd.Flags().StringVar(&ConfigData.Build, "build", "build-aux/setup.sh", "构建命令")
	cmd.Flags().BoolVar(&CreateFlag.NoBuild, "no-build", false, "不自动初始化项目")
	cmd.Flags().BoolVar(&CreateFlag.Force, "force", false, "强制覆盖已存在文件")
	cmd.Flags().StringVar(&CreateFlag.Metadata, "from", "", "从APT Package元数据创建(支持apt show)")

	return cmd
}
