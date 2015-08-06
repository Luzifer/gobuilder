package builddbCreator

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Luzifer/gobuilder/builddb"
)

// GenerateBuildDB is a helper to add new labels and files for those labels
// to an existing BuildDB
func GenerateBuildDB(basedir string) error {
	var buildDB builddb.BuildDB

	content, err := ioutil.ReadFile(fmt.Sprintf("%s/.build.db", basedir))
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &buildDB)
	if err != nil {
		return err
	}

	goBuildVersion, err := ioutil.ReadFile(fmt.Sprintf("%s/.goversion", basedir))
	if err != nil {
		return err
	}

	cache := make(map[string]map[string]os.FileInfo)
	files, _ := ioutil.ReadDir(fmt.Sprintf("%s/", basedir))

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".zip") {
			tmp := strings.Split(f.Name(), "_")
			buildName := tmp[len(tmp)-2]

			if _, ok := cache[buildName]; !ok {
				cache[buildName] = make(map[string]os.FileInfo)
			}

			cache[buildName][f.Name()] = f
		}
	}

	for branch, fileNames := range cache {
		tmp := builddb.Branch{
			GoVersion: string(goBuildVersion),
			BuildDate: time.Now(),
			Assets:    []builddb.Asset{},
		}

		for _, f := range fileNames {
			md5sum, sha1sum, sha256sum := buildHashes(f.Name())
			tmp.Assets = append(tmp.Assets, builddb.Asset{
				Size:     f.Size(),
				SHA1:     sha1sum,
				SHA256:   sha256sum,
				MD5:      md5sum,
				FileName: f.Name(),
			})
		}

		sort.Sort(builddb.ByFilename(tmp.Assets))

		buildDB[branch] = tmp
	}

	db, err := json.Marshal(buildDB)
	ioutil.WriteFile(fmt.Sprintf("%s/.build.db", basedir), db, 0664)

	return nil
}

func buildHashes(filename string) (string, string, string) {
	fileContent, _ := ioutil.ReadFile(filename)
	sha1sum := fmt.Sprintf("%x", sha1.Sum(fileContent))
	sha256sum := fmt.Sprintf("%x", sha256.Sum256(fileContent))
	md5sum := fmt.Sprintf("%x", md5.Sum(fileContent))
	return md5sum, sha1sum, sha256sum
}
