package appcfg

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Dirs struct {
	Data       string
	Runtime    string
	Node       string
	Dsh        string
	Home       string
	Logs       string
	NpmCache   string
	ConfigFile string
}

type Config struct {
	DSHVersion string `json:"dsh_version"`
	Registry   string `json:"registry"`
	Dirs       Dirs   `json:"-"`
	Portable   bool   `json:"-"`
}

func Load() (*Config, error) {
	exe, err := os.Executable()
	if err != nil {
		exe = "."
	}
	exeDir := filepath.Dir(exe)
	portable := fileExists(filepath.Join(exeDir, "portable.marker"))

	var data string
	if portable {
		data = exeDir
	} else {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = os.TempDir()
		}
		data = filepath.Join(base, "my-dsh-desktop")
	}

	var runtime string
	switch {
	case portable:
		runtime = filepath.Join(exeDir, "runtime")
	case dirExists(filepath.Join(exeDir, "runtime")):
		runtime = filepath.Join(exeDir, "runtime")
	default:
		runtime = filepath.Join(data, "runtime")
	}

	c := &Config{
		DSHVersion: "0.1.0-rc.6",
		Portable:   portable,
		Dirs: Dirs{
			Data:       data,
			Runtime:    runtime,
			Node:       filepath.Join(runtime, "node"),
			Dsh:        filepath.Join(runtime, "dsh"),
			Home:       filepath.Join(data, "home"),
			Logs:       filepath.Join(data, "logs"),
			NpmCache:   filepath.Join(data, "npm-cache"),
			ConfigFile: filepath.Join(data, "config.json"),
		},
	}

	if b, err := os.ReadFile(c.Dirs.ConfigFile); err == nil {
		var saved struct {
			DSHVersion string `json:"dsh_version"`
			Registry   string `json:"registry"`
		}
		if json.Unmarshal(b, &saved) == nil {
			if saved.DSHVersion != "" {
				c.DSHVersion = saved.DSHVersion
			}
			c.Registry = saved.Registry
		}
	}
	return c, nil
}

func (c *Config) EnsureDirs() error {
	for _, d := range []string{c.Dirs.Data, c.Dirs.Home, c.Dirs.Logs, c.Dirs.NpmCache} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}
