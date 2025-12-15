package service

import (
	"bufio"
	"os"
)

type FileProducer struct {
	filePath string
}

func NewFileProducer(path string) *FileProducer {
	return &FileProducer{filePath: path}
}

func (producer *FileProducer) Produce() ([]string, error) {
	file, err := os.Open(producer.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
