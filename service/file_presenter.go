package service

import (
	"os"
	"strings"
)

type FilePresenter struct {
	filePath string
}

func NewFilePresenter(path string) *FilePresenter {
	if path == "" {
		path = "C:/GoLand/GoCourse/LinkMaskirator/txtFiles/maskedLinks.txt"
	}
	return &FilePresenter{filePath: path}
}

func trimSpaces(lines []string) string {
	var trimmed []string
	for _, item := range lines {
		if strings.TrimSpace(item) != "" {
			trimmed = append(trimmed, item)
		}
	}
	return strings.Join(trimmed, "\n")
}

func (presenter *FilePresenter) Present(lines []string) error {
	data := trimSpaces(lines)
	return os.WriteFile(presenter.filePath, []byte(data), 0644)

}
