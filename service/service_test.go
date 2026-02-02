package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProducer - мок для интерфейса Producer
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Produce() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockPresenter - мок для интерфейса Presenter
type MockPresenter struct {
	mock.Mock
}

func (m *MockPresenter) Present(lines []string) error {
	args := m.Called(lines)
	return args.Error(0)
}

// / TestNewServiceFactory - тест фабрики (добавляем slowmode)
func TestNewServiceFactory(t *testing.T) {
	t.Run("создание фабрики с workers и slowmode=false", func(t *testing.T) {
		factory := NewServiceFactory(5, false)
		assert.Equal(t, 5, factory._workers)
		assert.False(t, factory._slowmode)
	})

	t.Run("создание фабрики с workers и slowmode=true", func(t *testing.T) {
		factory := NewServiceFactory(3, true)
		assert.Equal(t, 3, factory._workers)
		assert.True(t, factory._slowmode)
	})

	t.Run("создание фабрики с дефолтными workers", func(t *testing.T) {
		factory := NewServiceFactory(0, false)
		assert.NotNil(t, factory)
		assert.Equal(t, 0, factory._workers)
		assert.False(t, factory._slowmode)
	})
}

// TestServiceFactory_CreateMaskService - тест создания сервиса с slowmode
func TestServiceFactory_CreateMaskService(t *testing.T) {
	// Тест 1: создание сервиса без slowmode
	t.Run("создание сервиса без slowmode", func(t *testing.T) {
		factory := NewServiceFactory(10, false)
		service := factory.CreateMaskService("input.txt", "output.txt")

		assert.NotNil(t, service)
		assert.Equal(t, 10, service.GetWorkers())
		assert.False(t, service.CheckSlowMode())
	})

	// Тест 2: создание сервиса с slowmode
	t.Run("создание сервиса с slowmode", func(t *testing.T) {
		factory := NewServiceFactory(7, true)
		service := factory.CreateMaskService("input.txt", "output.txt")

		assert.NotNil(t, service)
		assert.Equal(t, 7, service.GetWorkers())
		assert.True(t, service.CheckSlowMode())
	})

	// Тест 3: сервис с кастомными зависимостями и slowmode
	t.Run("сервис с кастомными slowmode настройками", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		service := NewService(mockProducer, mockPresenter)
		service.SetWorkers(5)

		// Проверяем что slowmode по умолчанию false
		assert.False(t, service.CheckSlowMode())

		// Включаем slowmode
		service.SetSlowMode(true)
		assert.True(t, service.CheckSlowMode())

		// Выключаем slowmode
		service.SetSlowMode(false)
		assert.False(t, service.CheckSlowMode())
	})
}

// TestService_RunWithContext - основные тесты с контекстом (добавляем slowmode)
func TestService_RunWithContext(t *testing.T) {
	t.Run("успешное выполнение с slowmode=false", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		inputLines := []string{
			"текст с http://ссылкой",
			"еще текст https://example.com",
		}
		expectedOutput := []string{
			"текст с http://*******",
			"еще текст https://*********",
		}

		mockProducer.On("Produce").Return(inputLines, nil)
		mockPresenter.On("Present", expectedOutput).Return(nil)

		service := NewService(mockProducer, mockPresenter)
		service.SetWorkers(2)
		service.SetSlowMode(false) // Явно выключаем slowmode

		ctx := context.Background()
		err := service.Run(ctx)

		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
		mockPresenter.AssertExpectations(t)
	})

	// Новый тест: проверяем что slowmode не влияет на результат
	t.Run("результат одинаковый с slowmode=true и false", func(t *testing.T) {
		mockProducer1 := new(MockProducer)
		mockPresenter1 := new(MockPresenter)
		mockProducer2 := new(MockProducer)
		mockPresenter2 := new(MockPresenter)

		inputLines := []string{
			"http://example.com",
			"https://site.org/path",
		}
		expectedOutput := []string{
			"http://*********",
			"https://*************",
		}

		// Настраиваем первый сервис (без slowmode)
		mockProducer1.On("Produce").Return(inputLines, nil)
		mockPresenter1.On("Present", expectedOutput).Return(nil)

		service1 := NewService(mockProducer1, mockPresenter1)
		service1.SetWorkers(2)
		service1.SetSlowMode(false)

		// Настраиваем второй сервис (с slowmode)
		mockProducer2.On("Produce").Return(inputLines, nil)
		mockPresenter2.On("Present", expectedOutput).Return(nil)

		service2 := NewService(mockProducer2, mockPresenter2)
		service2.SetWorkers(2)
		service2.SetSlowMode(true)

		// Запускаем оба
		ctx := context.Background()
		err1 := service1.Run(ctx)
		err2 := service2.Run(ctx)

		// Результат должен быть одинаковым
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		mockPresenter1.AssertExpectations(t)
		mockPresenter2.AssertExpectations(t)
	})

	t.Run("slowmode=true с быстрым таймаутом (должен успеть)", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		inputLines := []string{"http://test.com"}
		mockProducer.On("Produce").Return(inputLines, nil)
		mockPresenter.On("Present", mock.Anything).Return(nil)

		service := NewService(mockProducer, mockPresenter)
		service.SetWorkers(1)
		service.SetSlowMode(true) // Slowmode включен

		// Таймаут достаточный даже с slowmode
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := service.Run(ctx)

		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
		mockPresenter.AssertExpectations(t)
	})

	t.Run("slowmode=true с таймаутом контекста (должен прерваться)", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		inputLines := []string{"строка1", "строка2", "строка3"}
		mockProducer.On("Produce").Return(inputLines, nil)

		service := NewService(mockProducer, mockPresenter)
		service.SetWorkers(1)
		service.SetSlowMode(true) // Включаем slowmode

		// Очень короткий таймаут
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Запускаем сервис с замедлением
		err := service.Run(ctx)

		// Должен быть таймаут
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)

		// Презентер не должен вызываться
		mockPresenter.AssertNotCalled(t, "Present")
	})

	t.Run("отмена контекста при slowmode=true", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		inputLines := []string{"http://example.com"}
		mockProducer.On("Produce").Return(inputLines, nil)

		service := NewService(mockProducer, mockPresenter)
		service.SetWorkers(1)
		service.SetSlowMode(true)

		ctx, cancel := context.WithCancel(context.Background())

		// Отменяем почти сразу
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		err := service.Run(ctx)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		mockPresenter.AssertNotCalled(t, "Present")
	})
}

// TestService_ConcurrentProcessing - тесты конкурентной обработки с slowmode
func TestService_ConcurrentProcessing(t *testing.T) {
	t.Run("обработка с slowmode=true и большим количеством workers", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		inputLines := []string{
			"http://link1.com",
			"https://link2.org",
			"http://link3.net",
		}

		mockProducer.On("Produce").Return(inputLines, nil)
		mockPresenter.On("Present", mock.AnythingOfType("[]string")).Return(nil)

		service := NewService(mockProducer, mockPresenter)
		service.SetWorkers(5)     // Больше чем строк
		service.SetSlowMode(true) // Включаем slowmode

		// Даем больше времени из-за slowmode
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		err := service.Run(ctx)

		assert.NoError(t, err)
		mockPresenter.AssertCalled(t, "Present", mock.MatchedBy(func(lines []string) bool {
			return len(lines) == 3
		}))
	})

	t.Run("slowmode не влияет на корректность маскировки", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{"простая ссылка", "http://example.com", "http://*********"},
			{"ссылка в тексте", "текст http://link.ru текст", "текст http://******* текст"},
		}

		var inputs []string
		var expected []string

		for _, tc := range testCases {
			inputs = append(inputs, tc.input)
			expected = append(expected, tc.expected)
		}

		mockProducer.On("Produce").Return(inputs, nil)
		mockPresenter.On("Present", expected).Return(nil)

		service := NewService(mockProducer, mockPresenter)
		service.SetSlowMode(true) // Включаем slowmode

		err := service.Run(context.Background())

		require.NoError(t, err)
		mockPresenter.AssertExpectations(t)
	})
}

// TestService_SetSlowMode - тесты настройки slowmode
func TestService_SetSlowMode(t *testing.T) {
	t.Run("включение и выключение slowmode", func(t *testing.T) {
		service := NewService(nil, nil)

		// По умолчанию должен быть false
		assert.False(t, service.CheckSlowMode())

		// Включаем
		service.SetSlowMode(true)
		assert.True(t, service.CheckSlowMode())

		// Выключаем
		service.SetSlowMode(false)
		assert.False(t, service.CheckSlowMode())

		// Снова включаем
		service.SetSlowMode(true)
		assert.True(t, service.CheckSlowMode())
	})

	t.Run("slowmode с разными значениями workers", func(t *testing.T) {
		service := NewService(nil, nil)

		service.SetWorkers(1)
		service.SetSlowMode(true)
		assert.True(t, service.CheckSlowMode())
		assert.Equal(t, 1, service.GetWorkers())

		service.SetWorkers(10)
		assert.True(t, service.CheckSlowMode()) // slowmode не должен сброситься
		assert.Equal(t, 10, service.GetWorkers())
	})
}

// TestService_WorkerSlowMode - тесты поведения worker с slowmode
func TestService_WorkerSlowMode(t *testing.T) {
	t.Run("worker с slowmode=true добавляет задержку", func(t *testing.T) {
		// Используем реальные зависимости для интеграционного теста
		producer := &MockProducer{}
		presenter := &MockPresenter{}

		inputLines := []string{"http://test1.com", "http://test2.com"}
		producer.On("Produce").Return(inputLines, nil)
		presenter.On("Present", mock.Anything).Return(nil)

		service := NewService(producer, presenter)
		service.SetWorkers(1)
		service.SetSlowMode(true)

		// Замеряем время выполнения
		start := time.Now()
		err := service.Run(context.Background())
		elapsed := time.Since(start)

		assert.NoError(t, err)

		// С slowmode должно занять минимум 200ms (2 строки × 100ms)
		// Плюс время на обработку
		minExpected := 200 * time.Millisecond
		assert.GreaterOrEqual(t, elapsed, minExpected,
			"с slowmode=true выполнение должно занимать больше времени")

		t.Logf("Время выполнения с slowmode=true: %v", elapsed)
	})

	t.Run("worker с slowmode=false выполняется быстро", func(t *testing.T) {
		producer := &MockProducer{}
		presenter := &MockPresenter{}

		inputLines := []string{"http://test1.com", "http://test2.com", "http://test3.com"}
		producer.On("Produce").Return(inputLines, nil)
		presenter.On("Present", mock.Anything).Return(nil)

		service := NewService(producer, presenter)
		service.SetWorkers(1)
		service.SetSlowMode(false)

		start := time.Now()
		err := service.Run(context.Background())
		elapsed := time.Since(start)

		assert.NoError(t, err)

		// Без slowmode должно быть быстро
		maxExpected := 100 * time.Millisecond
		assert.Less(t, elapsed, maxExpected,
			"с slowmode=false выполнение должно быть быстрым")

		t.Logf("Время выполнения с slowmode=false: %v", elapsed)
	})
}

// TestWorkerPool - интеграционные тесты пула воркеров (обновляем)
func TestWorkerPool(t *testing.T) {
	t.Run("обработка большой нагрузки с slowmode=false", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		var inputLines []string
		for i := 0; i < 100; i++ { // Уменьшим для скорости тестов
			inputLines = append(inputLines, "http://example.com")
		}

		mockProducer.On("Produce").Return(inputLines, nil)
		mockPresenter.On("Present", mock.MatchedBy(func(lines []string) bool {
			return len(lines) == 100
		})).Return(nil)

		service := NewService(mockProducer, mockPresenter)
		service.SetWorkers(10)
		service.SetSlowMode(false)

		err := service.Run(context.Background())

		assert.NoError(t, err)
		mockPresenter.AssertExpectations(t)
	})

	t.Run("обработка с slowmode=true и несколькими workers", func(t *testing.T) {
		mockProducer := new(MockProducer)
		mockPresenter := new(MockPresenter)

		// Маленький набор для теста скорости
		inputLines := []string{"line1", "line2", "line3"}
		mockProducer.On("Produce").Return(inputLines, nil)
		mockPresenter.On("Present", mock.Anything).Return(nil)

		service := NewService(mockProducer, mockPresenter)
		service.SetWorkers(3)
		service.SetSlowMode(true)

		// Даем больше времени из-за slowmode
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := service.Run(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err)
		// С 3 строками и slowmode должно быть минимум 300ms
		assert.GreaterOrEqual(t, elapsed, 300*time.Millisecond)
		t.Logf("Время с 3 workers и slowmode: %v", elapsed)
	})
}

// Benchmark тесты с slowmode
func BenchmarkService_Run_WithSlowMode(b *testing.B) {
	producer := &MockProducer{}
	presenter := &MockPresenter{}

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "http://example.com"
	}

	producer.On("Produce").Return(lines, nil)
	presenter.On("Present", mock.Anything).Return(nil)

	// Тест с slowmode=true
	b.Run("slowmode=true", func(b *testing.B) {
		service := NewService(producer, presenter)
		service.SetWorkers(10)
		service.SetSlowMode(true)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = service.Run(context.Background())
		}
	})

	// Тест с slowmode=false
	b.Run("slowmode=false", func(b *testing.B) {
		service := NewService(producer, presenter)
		service.SetWorkers(10)
		service.SetSlowMode(false)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = service.Run(context.Background())
		}
	})
}
