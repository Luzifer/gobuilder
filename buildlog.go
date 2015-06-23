package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

func handlerBuildLog(res http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	file, err := redisClient.Get(fmt.Sprintf("project::%s::logs::%s", params["repo"], params["logid"]))
	if err != nil {
		file = []byte("No build log was found for this build.")
	}

	template := pongo2.Must(pongo2.FromFile("frontend/buildlog.html"))
	ctx := getBasicContext(r)
	ctx["repo"] = params["repo"]
	ctx["log"] = logHighlight(file)

	template.ExecuteWriter(ctx, res)

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
