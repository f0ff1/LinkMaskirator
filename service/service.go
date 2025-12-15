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
	result := []byte(message)
	httpString := "http://"

	for i := 0; i < len(result)-len(httpString); i++ {
		if string(result[i:i+len(httpString)]) == httpString {
			startHttp := i + len(httpString)
			for j := startHttp; j < len(result); j++ {
				if result[j] == ' ' {
					break
				}
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

	maskedLines := make([]string, len(data))

	for _, line := range data {
		maskedLines = append(maskedLines, maskLink(line))
	}

	err = s._pres.Present(maskedLines)
	if err != nil {
		return err
	}
	return nil

}
