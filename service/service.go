package service

import (
	"fmt"
)

type Producer interface {
	Produce() ([]string, error)
}

type Presenter interface {
	Present([]string) error
}

type Service struct {
	_prod Producer
	_pres Presenter
}

func NewService(prod Producer, pres Presenter) *Service {
	return &Service{_prod: prod, _pres: pres}
}

func maskLink(message string) string {
	result := []rune(message)
	httpString := "http://"

	httpRunes := []rune(httpString)

	for i := 0; i <= len(result)-len(httpRunes); i++ {
		found := true
		for j := 0; j < len(httpRunes); j++ {
			if result[i+j] != httpRunes[j] {
				found = false
				break
			}
		}

		if found {
			startMask := i + len(httpRunes)
			for j := startMask; j < len(result) && result[j] != ' '; j++ {
				result[j] = '*'
			}
		}

	}
	return string(result)
}

func (s *Service) Run() error {
	data, err := s._prod.Produce()
	if err != nil {
		return err
	}

	for _, item := range data {
		fmt.Println(item)
	}

	maskedLines := make([]string, 0, len(data))

	for _, line := range data {
		maskedLines = append(maskedLines, maskLink(line))
	}

	err = s._pres.Present(maskedLines)
	if err != nil {
		return err
	}
	return nil

}
