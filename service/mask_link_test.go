package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskLink(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "обычная ссылка",
			input:    "Посетите http://example.com сейчас",
			expected: "Посетите http://*********** сейчас",
		},
		{
			name:     "две ссылки",
			input:    "Первая http://one.com и вторая http://two.com",
			expected: "Первая http://******* и вторая http://*******",
		},
		{
			name:     "без ссылок",
			input:    "Просто текст",
			expected: "Просто текст",
		},
		{
			name:     "ссылка в начале",
			input:    "http://site.com текст",
			expected: "http://******** текст",
		},
		{
			name:     "ссылка в конце",
			input:    "Текст http://link",
			expected: "Текст http://****",
		},
		{
			name:     "пустая строка",
			input:    "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := maskLink(test.input)
			assert.Equal(t, test.expected, result,
				"Маскировка не сработала для: %d", test.name)
		})
	}
}
