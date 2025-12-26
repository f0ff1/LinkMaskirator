package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimSpace_Succes(t *testing.T) {

	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"Пустой слайс", []string{}, ""},
		{"Только пробелы", []string{" ", "\t", "\n"}, ""},
		{"Микс всего", []string{" a ", "b  ", "  c"}, "a\nb\nc"},
		{"Вообще без пробелов", []string{"hello", "world"}, "hello\nworld"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := trimSpaces(test.input)
			assert.Equal(t, test.expected, result)
		})
	}

}
