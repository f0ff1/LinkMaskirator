package service

import (
	"context"
	"strings"
	"time"
	// "sync"
)

type Producer interface {
	Produce() ([]string, error)
}

type Presenter interface {
	Present([]string) error
}

type Service struct {
	_prod     Producer
	_pres     Presenter
	_workers  int
	_slowmode bool
}

func NewService(prod Producer, pres Presenter) *Service {
	return &Service{
		_prod:     prod,
		_pres:     pres,
		_workers:  10,
		_slowmode: false,
	}
}

func (s *Service) SetWorkers(count int) {
	if count > 0 {
		s._workers = count
	}
}

func (s *Service) GetWorkers() int {
	return s._workers
}

func (s *Service) SetSlowMode(enabled bool) {
	s._slowmode = enabled
}

func (s *Service) CheckSlowMode() bool {
	return s._slowmode
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

func (s *Service) Run(ctx context.Context) error {
	data, err := s._prod.Produce()
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	workersCount := s.GetWorkers()
	origLinesChan := make(chan string)
	resultLinesChan := make(chan string)

	// А что если у нас меньше строк? Нафига тогда 10 воркеров?
	if workersCount > len(data) {
		workersCount = len(data)
	}
	// var wg sync.WaitGroup

	for i := 0; i < workersCount; i++ {
		// wg.Add(1)
		go s.Worker(ctx, origLinesChan, resultLinesChan)
	}

	// Отправляю строки для маскировки в канал. Закрываю по завершению цикла.
	go func() {
		defer close(origLinesChan)
		for _, line := range data {
			select {
			case <-ctx.Done():
				return
			default:
				origLinesChan <- line
			}

		}

	}()

	maskedLines := make([]string, 0, len(data))

	for i := 0; i < len(data); i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case resultLine := <-resultLinesChan:
			maskedLines = append(maskedLines, resultLine)
		}

	}

	// wg.Wait()
	close(resultLinesChan)

	err = s._pres.Present(maskedLines)
	if err != nil {
		return err
	}
	return nil

}

func (s *Service) Worker(ctx context.Context, origLinesChan <-chan string, resultLinesChan chan<- string) {
	// defer wg.Done()
	isSlowMode := s.CheckSlowMode()
	for orLine := range origLinesChan {
		select {
		case <-ctx.Done():
			return
		default:
			if isSlowMode {
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return
				}
			}
			resultLinesChan <- maskLink(orLine)

		}

	}

}
