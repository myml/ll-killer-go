/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package types

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
