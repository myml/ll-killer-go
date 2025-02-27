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

	"github.com/go-yaml/yaml"
	"github.com/urfave/cli/v2"
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
	NoBuild   bool
	Metadata  string
	Extend    string
	FieldMask map[string]bool
}

func NormalizeVersion(version string) string {
	re := regexp.MustCompile(`[^\d\.]+`)
	chunks := strings.SplitN(version, ".", 4)
	for index, chunk := range chunks {
		chunks[index] = strings.TrimLeft(strings.TrimSpace(re.ReplaceAllString(chunk, "")), "0")
	}
	for len(chunks) < 4 {
		chunks = append(chunks, "0")
	}
	return strings.Join(chunks, ".")
}
func SetupPackageMetadata(cmd *cli.Command) error {
	metadata, err := ParsePackageMetadataFromFile(CreateFlag.Metadata)
	if err != nil {
		return err
	}
	if !CreateFlag.FieldMask["description"] && metadata["description"] != "" {
		ConfigData.Package.Description = metadata["description"]
	}
	if !CreateFlag.FieldMask["version"] && metadata["version"] != "" {
		ConfigData.Package.Version = NormalizeVersion(metadata["version"])
	}
	if !CreateFlag.FieldMask["id"] && metadata["package"] != "" {
		ConfigData.Package.ID = metadata["package"]
	}
	if !CreateFlag.FieldMask["name"] && metadata["package"] != "" {
		ConfigData.Package.Name = metadata["package"]
	}
	if !CreateFlag.FieldMask["base"] && metadata["base"] != "" {
		ConfigData.Package.Name = metadata["runtime"]
	}
	if !CreateFlag.FieldMask["runtime"] && metadata["runtime"] != "" {
		ConfigData.Package.Name = metadata["runtime"]
	}
	if metadata["apt-sources"] != "" {
		if !IsExist(SourceListFile) {
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
				log.Printf("created: %s\n", SourceListFile)
				err := WriteFile(SourceListFile, []byte(strings.Join(parsed, "\n")), 0755)
				if err != nil {
					return err
				}
			}
		} else {
			log.Printf("skip: %s\n", SourceListFile)
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

func CreateMain(ctx *cli.Context) error {
	err := embedFilesToDisk(".")
	if err != nil {
		return err
	}

	SetupPackageMetadata(ctx.Command)

	fs, err := os.Create(kLinglongYaml)
	if err != nil {
		return err
	}
	encoder := yaml.NewEncoder(fs)
	err = encoder.Encode(ConfigData)
	if err != nil {
		return err
	}
	err = encoder.Close()
	if err != nil {
		return err
	}
	log.Printf("created: %s\n", kLinglongYaml)
	if !CreateFlag.NoBuild {
		return RunCommand("ll-builder", "build", "--exec", "true")
	}
	return nil
}
func SetMask[T any](name string) func(ctx *cli.Context, v T) error {
	return func(ctx *cli.Context, v T) error {
		CreateFlag.FieldMask[name] = true
		return nil
	}
}
func CreateCreateCommand() *cli.Command {
	CreateFlag.FieldMask = make(map[string]bool)
	return &cli.Command{
		Name:  "create",
		Usage: "创建模板项目",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "spec",
				Usage:       "玲珑yaml版本",
				Value:       "1",
				Destination: &ConfigData.Version,
				Action:      SetMask[string]("spec"),
			},
			&cli.StringFlag{
				Name:        "id",
				Usage:       "包名",
				Value:       "app",
				Destination: &ConfigData.Package.ID,
				Action:      SetMask[string]("id"),
			},
			&cli.StringFlag{
				Name:        "name",
				Usage:       "显示名称",
				Value:       "app",
				Destination: &ConfigData.Package.Name,
				Action:      SetMask[string]("name"),
			},
			&cli.StringFlag{
				Name:        "version",
				Usage:       "版本号",
				Value:       "0.0.0.1",
				Destination: &ConfigData.Package.Version,
				Action:      SetMask[string]("version"),
			},
			&cli.StringFlag{
				Name:        "kind",
				Usage:       "应用类型：app|runtime",
				Value:       "app",
				Destination: &ConfigData.Package.Kind,
				Action:      SetMask[string]("kind"),
			},
			&cli.StringFlag{
				Name:        "description",
				Usage:       "应用说明",
				Value:       "",
				Destination: &ConfigData.Package.Description,
				Action:      SetMask[string]("description"),
			},
			&cli.MultiStringFlag{
				Target: &cli.StringSliceFlag{
					Name:  "command",
					Usage: "启动命令",
				},
				Value:       []string{"entrypoint.sh"},
				Destination: &ConfigData.Command,
			},
			&cli.StringFlag{
				Name:        "base",
				Usage:       "Base镜像",
				Value:       "org.deepin.base/23.1.0",
				Destination: &ConfigData.Base,
				Action:      SetMask[string]("base"),
			},
			&cli.StringFlag{
				Name:        "runtime",
				Usage:       "Runtime镜像",
				Value:       "",
				Destination: &ConfigData.Runtime,
				Action:      SetMask[string]("runtime"),
			},
			&cli.StringFlag{
				Name:        "build",
				Usage:       "构建命令",
				Value:       "build-aux/setup.sh",
				Destination: &ConfigData.Build,
				Action:      SetMask[string]("build"),
			},
			&cli.BoolFlag{
				Name:        "no-build",
				Usage:       "不自动初始化项目",
				Value:       false,
				Destination: &CreateFlag.NoBuild,
				Action:      SetMask[bool]("no-build"),
			},
			&cli.StringFlag{
				Name:        "from",
				Usage:       "从APT Package元数据创建(支持apt show)",
				Destination: &CreateFlag.Metadata,
				TakesFile:   true,
				Action:      SetMask[string]("from"),
			},
		},
		Action: CreateMain,
	}
}
