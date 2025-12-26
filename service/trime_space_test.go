package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

)

func TestTrimSpace_Succes(t *testing.T) {
	testedLines := []string{"Как дела", "", "Хорошо", " ", "Хелло"}
	expected := "Как дела\nХорошо\nХелло"

	result := trimSpaces(testedLines)
	assert.Equal(t, expected, result)

}
