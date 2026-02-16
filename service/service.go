package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
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
	var wg sync.WaitGroup

	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go s.Worker(ctx, origLinesChan, resultLinesChan, &wg)
	}

	// Отправляю строки для маскировки в канал.
	go func() {
		defer close(origLinesChan)
		for _, line := range data {
			select {
			case <-ctx.Done():
				slog.DebugContext(ctx, "прекращена отправка данных для маскировки")
				return
			default:
				origLinesChan <- line
			}

		}

	}()

	maskedLines := make([]string, 0, len(data))
	var collectWg sync.WaitGroup
	collectWg.Add(1)

	go func() {
		defer collectWg.Done()
		for i := 0; i < len(data); i++ {
			select {
			case <-ctx.Done():
				slog.DebugContext(ctx, "прекращается отправка замаскированных данных")
				return
			case resultLine := <-resultLinesChan:
				maskedLines = append(maskedLines, resultLine)
			}
		}
	}()

	wg.Wait()
	close(resultLinesChan)

	collectWg.Wait()

	if len(maskedLines) > 0 {
		if err := s._pres.Present(maskedLines); err != nil {
			slog.DebugContext(ctx, "ошибка сохранения данных в файл", "error", err)
			return fmt.Errorf("ошибка сохранения: %w", err)
		}
		slog.InfoContext(ctx, "результаты сохранены",
			"lines_saved", len(maskedLines),
			"total", len(data))
	}

	if ctx.Err() != nil {
		slog.InfoContext(ctx, "обработка завершена по сигналу",
			"reason", ctx.Err())
		return ctx.Err()
	}

	return nil

}

func (s *Service) Worker(ctx context.Context, origLinesChan <-chan string, resultLinesChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
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
