package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/Luzifer/gobuilder/builddb"
)

type noDownloadFoundError struct{}
type generalDownloadError struct {
	Message string
}

func (n noDownloadFoundError) Error() string { return "No download was found." }
func (g generalDownloadError) Error() string { return g.Message }

func downloadBuildResult(repo, version, search, target string) error {
	resp, err := http.Get(fmt.Sprintf("https://gobuilder.me/api/v1/%s/hashes/%s.json", repo, version))
	if err != nil {
		return generalDownloadError{"Was unable to communicate with GoBuilder, please try again."}
	}
	defer resp.Body.Close()

	hdb := builddb.HashDB{}
	if err := json.NewDecoder(resp.Body).Decode(&hdb); err != nil {
		return generalDownloadError{fmt.Sprintf("Unable to read hash-database: %s", err)}
	}

	downloaded := 0
	for k := range hdb {
		if strings.Contains(k, search) {
			if err := downloadAndCheck(repo, k, target, hdb[k]); err != nil {
				fmt.Printf("%s\n", err)
				continue
			}
			fmt.Printf("%s downloaded and verified successfully.\n", k)
			downloaded++
		}
	}

	if downloaded == 0 {
		return noDownloadFoundError{}
	}
	return nil
}

func downloadAndCheck(repo, filename, target string, hash builddb.Hashes) error {
	url := fmt.Sprintf("https://gobuilder.me/get/%s/%s", repo, filename)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("An error ocurred while downloading the package '%s': %s", filename, err)
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	s5, s1, s256 := buildHashes(body)

	if s5 != hash.MD5 || s1 != hash.SHA1 || s256 != hash.SHA256 {
		return fmt.Errorf("Download of %s failed, hashes did not match.", filename)
	}

	return ioutil.WriteFile(path.Join(target, filename), body, 0644)
}

func buildHashes(fileContent []byte) (string, string, string) {
	sha1sum := fmt.Sprintf("%x", sha1.Sum(fileContent))
	sha256sum := fmt.Sprintf("%x", sha256.Sum256(fileContent))
	md5sum := fmt.Sprintf("%x", md5.Sum(fileContent))
	return md5sum, sha1sum, sha256sum
}
