package codewriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteArithmetic(t *testing.T) {
	cw := CodeWriter{}
	err := cw.WriteArithmetic("invalid")
	assert.EqualError(t, err, "received invalid arithmetic operation: invalid")
}
