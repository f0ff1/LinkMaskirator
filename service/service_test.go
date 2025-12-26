package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProducer struct{ mock.Mock }
type MockPresenter struct{ mock.Mock }

func (m *MockProducer) Produce() ([]string, error) {
	return m.Called().Get(0).([]string), m.Called().Error(1)
}

func (m *MockPresenter) Present(lines []string) error {
	return m.Called(lines).Error(0)
}

func TestService_Run_Success(t *testing.T) {
	mockProducer := new(MockProducer)
	mockPresenter := new(MockPresenter)

	mockProducer.On("Produce").Return([]string{"текст с http://ссылкой"}, nil)
	mockPresenter.On("Present", []string{"текст с http://*******"}).Return(nil)

	service := NewService(mockProducer, mockPresenter)

	err := service.Run()

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)

}

func TestService_Run_ProduceError(t *testing.T) {
	mockProducer := new(MockProducer)
	mockPresenter := new(MockPresenter)

	mockProducer.On("Produce").Return([]string{}, errors.New("Файл не существует"))

	service := NewService(mockProducer, mockPresenter)

	err := service.Run()

	assert.Error(t, err)
	mockProducer.AssertExpectations(t)
	mockPresenter.AssertNotCalled(t, "Present")
}

func TestService_Run_PresentError(t *testing.T) {
	mockProducer := new(MockProducer)
	mockPresenter := new(MockPresenter)

	mockProducer.On("Produce").Return([]string{"текст с http://ссылкой"}, nil)
	mockPresenter.On("Present", mock.Anything).Return(errors.New("ошибка записи в файл"))

	service := NewService(mockProducer, mockPresenter)

	err := service.Run()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "записи в")

	mockProducer.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)

}
