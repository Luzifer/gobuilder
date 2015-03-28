package filters

import (
	"strings"

	"github.com/flosch/pongo2"
)

func init() {
	pongo2.RegisterFilter("is_mainarch", checkMainArch)
	pongo2.RegisterFilter("branchicon", getBranchIcon)
}

func checkMainArch(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	str := in.String()
	s := strings.Split(str, "-")
	if len(s) < 2 {
		return nil, &pongo2.Error{
			Sender:   "filter:branchicon",
			ErrorMsg: "Field did not contain a valid GoBuilder filename",
		}
	}
	s = strings.Split(s[len(s)-2], "_")

	for _, v := range []string{"linux", "darwin", "windows"} {
		if s[len(s)-1] == v {
			return pongo2.AsValue(true), nil
		}
	}
	return pongo2.AsValue(false), nil
}

func getBranchIcon(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	str := in.String()
	s := strings.Split(str, "-")
	if len(s) < 2 {
		return nil, &pongo2.Error{
			Sender:   "filter:branchicon",
			ErrorMsg: "Field did not contain a valid GoBuilder filename",
		}
	}
	s = strings.Split(s[len(s)-2], "_")

	// Map the architectures used by golang to fontawesome icon names
	switch s[len(s)-1] {
	case "darwin":
		return pongo2.AsValue("apple"), nil
	case "linux":
		return pongo2.AsValue("linux"), nil
	case "windows":
		return pongo2.AsValue("windows"), nil
	case "android":
		return pongo2.AsValue("android"), nil
	}

	// Not all archs have icons, use a generic file icon
	return pongo2.AsValue("file"), nil
}
