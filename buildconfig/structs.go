package buildconfig

import (
	"io/ioutil"

	"github.com/Luzifer/gobuilder/notifier"
	"gopkg.in/yaml.v2"
)

type BuildConfig struct {
	ReadmeFile  string                       `yaml:"readme_file,omitempty"`
	Artifacts   []string                     `yaml:"artifacts,omitempty"`
	Triggers    []string                     `yaml:"triggers,omitempty"`
	VersionFile string                       `yaml:"version_file,omitempty"`
	Notify      notifier.NotifyConfiguration `yaml:"notify,omitempty"`
	BuildMatrix map[string]ArchConfig        `yaml:"build_matrix,omitempty"`
}

type ArchConfig struct {
	Tags    []string `yaml:"build_tags"`
	LDFlags []string `yaml:"ldflags"`
}

func LoadFromFile(filepath string) (*BuildConfig, error) {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	tmp := &BuildConfig{
		BuildMatrix: make(map[string]ArchConfig),
	}
	err = yaml.Unmarshal(buf, tmp)
	if err != nil {
		return nil, err
	}

	return tmp, nil
}
