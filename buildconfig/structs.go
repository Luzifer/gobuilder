package buildconfig

import (
	"io/ioutil"

	"github.com/Luzifer/gobuilder/notifier"
	"gopkg.in/yaml.v2"
)

type BuildConfig struct {
	ReadmeFile  string                       `yaml:"readme_file,omitempty"`
	Artifacts   map[string]string            `yaml:"artifacts,omitempty"`
	Triggers    []string                     `yaml:"triggers,omitempty"`
	VersionFile string                       `yaml:"version_file,omitempty"`
	Notify      notifier.NotifyConfiguration `yaml:"notify,omitempty"`
	BuildMatrix map[string]ArchConfig        `yaml:"build_matrix,omitempty"`
}

type buildConfigV0 struct {
	ReadmeFile  string                       `yaml:"readme_file,omitempty"`
	Artifacts   []string                     `yaml:"artifacts,omitempty"`
	Triggers    []string                     `yaml:"triggers,omitempty"`
	VersionFile string                       `yaml:"version_file,omitempty"`
	Notify      notifier.NotifyConfiguration `yaml:"notify,omitempty"`
	BuildMatrix map[string]ArchConfig        `yaml:"build_matrix,omitempty"`
}

type buildConfigV1 BuildConfig

type ArchConfig struct {
	Tags    []string `yaml:"build_tags"`
	LDFlags []string `yaml:"ldflags"`
}

func LoadFromFile(filepath string) (*BuildConfig, error) {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	for {
		var t interface{}
		if err := yaml.Unmarshal(buf, &t); err != nil {
			return nil, err
		}

		if tmp, ok := t.(BuildConfig); ok {
			return &tmp, nil
		}

		if tmp0, ok := t.(buildConfigV0); ok {
			buf, err = upgradeConfigV0(tmp0)
		}
	}
}

func upgradeConfigV0(i buildConfigV0) ([]byte, error) {
	o := buildConfigV1{}
	o.BuildMatrix = i.BuildMatrix
	o.Notify = i.Notify
	o.ReadmeFile = i.ReadmeFile
	o.Triggers = i.Triggers
	o.VersionFile = i.VersionFile

	for _, v := range i.Artifacts {
		o.Artifacts[v] = v
	}

	return yaml.Marshal(o)
}
