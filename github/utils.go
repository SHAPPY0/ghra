package github

import (
	"strings"
	"strconv"
)

func GetRepoName(url string) string {
	repoName := url
	if url != "" {
		repo := url	
		if strings.HasSuffix(url, ".git") {
			repo = url[:len(url) - 4]
		}
		url_seg := strings.Split(repo, "/")
		repoName = url_seg[len(url_seg) - 1]
	}
	return repoName
}

func GetDepFilePath(path string) string {
	p := path
	if p != "" && p[0] == '/' {
		p = p[1:]
	}
	return p
} 

func StrToInt(val string) int {
	v, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return v
}