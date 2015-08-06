package builddb

import "time"

// The BuildDB is an archive for former builds
type BuildDB map[string]Branch

// Branch represents a "label" in the BuildDB
type Branch struct {
	GoVersion string    `json:"go_version"`
	BuildDate time.Time `json:"build_date"`
	Assets    []Asset   `json:"assets"`
}

// Asset contains information about the archive files in a BuildDBBranch
type Asset struct {
	SHA1     string `json:"sha1"`
	SHA256   string `json:"sha256"`
	MD5      string `json:"md5"`
	Size     int64  `json:"size"`
	FileName string `json:"file_name"`
}

// ByFilename implements a sorter for Assets
type ByFilename []Asset

func (f ByFilename) Len() int           { return len(f) }
func (f ByFilename) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByFilename) Less(i, j int) bool { return f[i].FileName < f[j].FileName }

// BranchSortEntry is a helper for the frontend to be able to sort the labels /
// branches by time
type BranchSortEntry struct {
	Branch    string
	BuildDate time.Time
}

// BranchSortEntryByBuildDate implements a sorter for BranchSortEntries
type BranchSortEntryByBuildDate []BranchSortEntry

func (b BranchSortEntryByBuildDate) Len() int           { return len(b) }
func (b BranchSortEntryByBuildDate) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b BranchSortEntryByBuildDate) Less(i, j int) bool { return b[i].BuildDate.Before(b[j].BuildDate) }

// HashDB is a database of file hashes
type HashDB map[string]Hashes

// Hashes contains the different hashes for a file in a HashDB
type Hashes struct {
	MD5    string `json:"md5sum" yaml:"md5sum"`
	SHA1   string `json:"sha1sum" yaml:"sha1sum"`
	SHA256 string `json:"sha256sum" yaml:"sha256sum"`
	SHA384 string `json:"sha384sum" yaml:"sha384sum"`
}
