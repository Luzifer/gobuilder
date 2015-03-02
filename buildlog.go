package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/go-martini/martini"
)

func handlerBuildLog(params martini.Params, res http.ResponseWriter, r *http.Request) {
	file, err := s3Bucket.Get(fmt.Sprintf("%s/build.log", params["repo"]))
	if err != nil {
		file = []byte("No build log was found for this build.")
	}

	template := pongo2.Must(pongo2.FromFile("frontend/buildlog.html"))
	template.ExecuteWriter(pongo2.Context{
		"repo": params["repo"],
		"log":  logHighlight(file),
	}, res)

}

type logline struct {
	Line         string
	BuildComment bool
}

func logHighlight(log []byte) []logline {
	lines := strings.Split(string(log), "\n")
	highlightedLines := []logline{}
	for _, line := range lines {
		tmp := logline{
			Line:         line,
			BuildComment: false,
		}
		if strings.HasPrefix(line, "[") {
			tmp.BuildComment = true
		}
		highlightedLines = append(highlightedLines, tmp)
	}
	return highlightedLines
}
