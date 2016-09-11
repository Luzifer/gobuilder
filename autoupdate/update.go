package autoupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Luzifer/gobuilder/builddb"
	"github.com/inconshreveable/go-update"
)

// Updater is a helper for autoupdating go-binaries with their builds on GoBuilder.me
type Updater struct {
	UpdateInterval time.Duration
	SelfRestart    bool

	repository        string
	label             string
	runningFile       string
	currentHash       string
	liveHash          string
	goBuilderFilename string
}

// New instanciates a new Updater
func New(repo, label string) *Updater {
	filename := fmt.Sprintf("%s_%s_%s-%s",
		path.Base(repo),
		label,
		runtime.GOOS,
		runtime.GOARCH,
	)

	if runtime.GOOS == "windows" {
		filename = filename + ".exe"
	}

	return &Updater{
		UpdateInterval:    time.Minute * 60,
		SelfRestart:       false,
		repository:        repo,
		label:             label,
		runningFile:       os.Args[0],
		goBuilderFilename: filename,
	}
}

// Run starts a permanent loop (better start as a go-routine) looking for updates
// to the current binary and will execute the update if required. This is intended
// for long running processes like daemons
func (g *Updater) Run() error {

	for range time.Tick(g.UpdateInterval) {
		if err := g.SingleRun(); err != nil {
			return err
		}
	}

	return nil
}

// SingleRun checks for an update and updates the binary when required.
func (g *Updater) SingleRun() error {
	bin, err := ioutil.ReadFile(g.runningFile)
	if err != nil {
		return err
	}

	g.currentHash = fmt.Sprintf("%x", sha256.Sum256(bin))

	liveHash, err := g.getGoBuilderHash()
	if err == nil && len(liveHash) == len(g.currentHash) && liveHash != g.currentHash {
		if err := g.updateBinary(); err == nil {
			if g.SelfRestart {
				args := []string{}
				if len(os.Args) > 1 {
					args = os.Args[1:]
				}
				syscall.Exec(os.Args[0], args, os.Environ())
			}
		} else {
			return fmt.Errorf("Update failed: %s", err)
		}
	}

	return nil
}

func (g *Updater) getGoBuilderHash() (string, error) {
	url := fmt.Sprintf("https://gobuilder.me/api/v1/%s/hashes/%s.json",
		g.repository,
		g.label,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP Status != 200")
	}

	out := builddb.HashDB{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	hashes, ok := out[g.goBuilderFilename]
	if !ok {
		return "", fmt.Errorf("Could not find hashes for %s", g.goBuilderFilename)
	}

	g.liveHash = hashes.SHA256

	return strings.TrimSpace(hashes.SHA256), nil
}

func (g *Updater) updateBinary() error {
	dlURL := fmt.Sprintf("https://gobuilder.me/get/%s/%s",
		g.repository,
		g.goBuilderFilename,
	)

	resp, err := http.Get(dlURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	cs, _ := hex.DecodeString(g.liveHash)

	return update.Apply(resp.Body, &update.Options{
		Checksum: cs,
	})
}
