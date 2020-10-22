package google

import (
	"regexp"
	"strings"
)

func GetResourceNameFromSelfLink(link string) string {
	parts := strings.Split(link, "/")
	return parts[len(parts)-1]
}

type LocationType int

const (
	Zonal LocationType = iota
	Regional
	Global
)

// return the region a selfLink is referring to
func GetRegionFromRegionSelfLink(selfLink string) string {
	re := regexp.MustCompile("/compute/[a-zA-Z0-9]*/projects/[a-zA-Z0-9-]*/regions/([a-zA-Z0-9-]*)")
	switch {
	case re.MatchString(selfLink):
		if res := re.FindStringSubmatch(selfLink); len(res) == 2 && res[1] != "" {
			return res[1]
		}
	}
	return selfLink
}
