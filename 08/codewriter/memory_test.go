package codewriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWritePushPop(t *testing.T) {
	cw := CodeWriter{}
	err := cw.WritePushPop(-1, "LCL", 1)

	assert.EqualError(t, err, "received invalid PushPop command: -1")
}

func TestPushDirect(t *testing.T) {
	cw := CodeWriter{}
	err := cw.pushDirect("INVALID", 1)
	assert.EqualError(t, err, "push: base address not found in map: INVALID")
}

func TestPopDirect(t *testing.T) {
	cw := CodeWriter{}
	err := cw.popDirect("INVALID", 1)
	assert.EqualError(t, err, "pop: base address not found in map: INVALID")
}
