package service

import (
	"strings"
	"sync"

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
	runes := []rune(message)
	lower := []rune(strings.ToLower(message))

	schemes := [][]rune{
		[]rune("http://"),
		[]rune("https://"),
	}

	for i := 0; i < len(lower); i++ {
		for _, scheme := range schemes {
			if i+len(scheme) > len(lower) {
				continue
			}
			match := true
			for j := range scheme {
				if lower[i+j] != scheme[j] {
					match = false
					break
				}
			}

			if !match {
				continue
			}

			start := i + len(scheme)
			for k := start; k < len(runes) && runes[k] != ' '; k++ {
				runes[k] = '*'
			}
		}

	}

	return string(runes)
}

func (s *Service) Run() error {
	data, err := s._prod.Produce()
	if err != nil {
		return err
	}

	workersCount := 10
	origLinesChan := make(chan string, workersCount)
	resultLinesChan := make(chan string, workersCount)

	// А что если у нас меньше строк? Нафига тогда 10 воркеров?
	if workersCount > len(data) {
		workersCount = len(data)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go Worker(origLinesChan, resultLinesChan, &wg)
	}

	// Отправляю строки для маскировки в канал. Закрываю по завершению цикла.
	go func() {
		defer close(origLinesChan)
		for _, line := range data {
			origLinesChan <- line
		}

	}()

	maskedLines := make([]string, 0, len(data))

	go func() {
		for resultLine := range resultLinesChan {
			mu.Lock()
			maskedLines = append(maskedLines, resultLine)
			mu.Unlock()
		}
	}()

	wg.Wait()
	close(resultLinesChan)

	err = s._pres.Present(maskedLines)
	if err != nil {
		return err
	}
	return nil

}

func Worker(origLinesChan <-chan string, resultLinesChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for orLine := range origLinesChan {
		resultLinesChan <- maskLink(orLine)
	}

}
