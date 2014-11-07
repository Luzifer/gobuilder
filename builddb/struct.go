package builddb

import "time"

type BuildDB map[string]BuildDBBranch

type BuildDBBranch struct {
	GoVersion string         `json:"go_version"`
	BuildDate time.Time      `json:"build_date"`
	Assets    []BuildDBAsset `json:"assets"`
}

type BuildDBAsset struct {
	SHA1     string `json:"sha1"`
	SHA256   string `json:"sha256"`
	MD5      string `json:"md5"`
	Size     int64  `json:"size"`
	FileName string `json:"file_name"`
}

type ByFilename []BuildDBAsset

func (f ByFilename) Len() int           { return len(f) }
func (f ByFilename) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByFilename) Less(i, j int) bool { return f[i].FileName < f[j].FileName }
