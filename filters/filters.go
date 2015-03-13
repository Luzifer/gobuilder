package filters

import (
	"strings"

	"github.com/flosch/pongo2"
)

func init() {
	pongo2.RegisterFilter("is_mainarch", checkMainArch)
}

func checkMainArch(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	str := in.String()
	s := strings.Split(strings.Split(str, "-")[0], "_")

	for _, v := range []string{"linux", "darwin", "windows"} {
		if s[len(s)-1] == v {
			return pongo2.AsValue(true), nil
		}
	}
	return pongo2.AsValue(false), nil
}
