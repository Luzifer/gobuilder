package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Luzifer/gobuilder/builddb"
)

func perror(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var buildDB builddb.BuildDB

	content, err := ioutil.ReadFile("build.db")
	perror(err)

	err = json.Unmarshal(content, &buildDB)
	perror(err)

	cache := make(map[string]map[string]os.FileInfo)
	files, _ := ioutil.ReadDir("./")

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

	for branch, files := range cache {
		buildDB[branch] = builddb.BuildDBBranch{
			GoVersion: os.Getenv("GO_VERSION"),
			BuildDate: time.Now(),
			Assets:    make(map[string]builddb.BuildDBAsset),
		}

		for _, f := range files {
			md5sum, sha1sum, sha256sum := buildHashes(f.Name())
			buildDB[branch].Assets[f.Name()] = builddb.BuildDBAsset{
				Size:   f.Size(),
				SHA1:   sha1sum,
				SHA256: sha256sum,
				MD5:    md5sum,
			}
		}
	}

	db, err := json.Marshal(buildDB)
	ioutil.WriteFile("build.db", db, 0664)
}

func buildHashes(filename string) (string, string, string) {
	fileContent, _ := ioutil.ReadFile(filename)
	sha1sum := fmt.Sprintf("%x", sha1.Sum(fileContent))
	sha256sum := fmt.Sprintf("%x", sha256.Sum256(fileContent))
	md5sum := fmt.Sprintf("%x", md5.Sum(fileContent))
	return md5sum, sha1sum, sha256sum
}
