package buildconfig

import (
	"fmt"
	"io/ioutil"

	"github.com/Luzifer/gobuilder/notifier"
	"gopkg.in/yaml.v2"
)

// BuildConfig represents a .gobuilder.yml file
type BuildConfig struct {
	ReadmeFile  string                       `yaml:"readme_file,omitempty"`
	Artifacts   map[string]string            `yaml:"artifacts,omitempty"`
	Triggers    []string                     `yaml:"triggers,omitempty"`
	VersionFile string                       `yaml:"version_file,omitempty"`
	Notify      notifier.NotifyConfiguration `yaml:"notify,omitempty"`
	BuildMatrix map[string]ArchConfig        `yaml:"build_matrix,omitempty"`
	NoGoFmt     string                       `yaml:"no_go_fmt,omitempty"`
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

// ArchConfig contains BuildTags and LDFlags for the specific architecture
type ArchConfig struct {
	Tags    []string `yaml:"build_tags"`
	LDFlags []string `yaml:"ldflags"`
}

// LoadFromFile retrieves the BuildConfig and transforms it into the latest
// config version if required
func LoadFromFile(filepath string) (*BuildConfig, error) {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	for {
		tmp := BuildConfig{}
		if err := yaml.Unmarshal(buf, &tmp); err == nil {
			return &tmp, nil
		}

		tmp0 := buildConfigV0{}
		if err := yaml.Unmarshal(buf, &tmp0); err == nil {
			buf, err = upgradeConfigV0(tmp0)
			continue
		}

		return nil, fmt.Errorf("Unable to parse BuildConfig")
	}
}

func upgradeConfigV0(i buildConfigV0) ([]byte, error) {
	o := buildConfigV1{
		Artifacts: make(map[string]string),
	}
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
