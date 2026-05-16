package codewriter

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type failWriter struct{}

func (failWriter) Write(p []byte) (int, error) { return 0, errors.New("write fail") }

func TestWriteLinesError(t *testing.T) {
	cw := &CodeWriter{writer: bufio.NewWriterSize(failWriter{}, 1)}
	err := cw.writeLines([]string{strings.Repeat("x", 100)})
	assert.Error(t, err)
}

func TestWriteLinesWriteByteError(t *testing.T) {
	cw := &CodeWriter{writer: bufio.NewWriterSize(failWriter{}, 1)}
	err := cw.writeLines([]string{"x"})
	assert.Error(t, err)
}

func TestWriteInitError(t *testing.T) {
	cw := &CodeWriter{writer: bufio.NewWriterSize(failWriter{}, 1)}
	err := cw.WriteInit()
	assert.Error(t, err)
}

func TestWriteInitWriteCallError(t *testing.T) {
	cw := &CodeWriter{writer: bufio.NewWriterSize(failWriter{}, 64)}
	err := cw.WriteInit()
	assert.Error(t, err)
}

func TestNewCodeWriterBootstrapError(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "x.asm")
	_, err := NewCodeWriter(tmp, true, WithWriter(bufio.NewWriterSize(failWriter{}, 1)))
	assert.Error(t, err)
}

func TestNewCodeWriterBadPath(t *testing.T) {
	_, err := NewCodeWriter("/no/such/dir/out.asm", false)
	assert.ErrorContains(t, err, "unable to create .asm file")
}

func TestWritePushPopInvalidCommand(t *testing.T) {
	cw := &CodeWriter{}
	err := cw.WritePushPop(999, "constant", 0)
	assert.ErrorContains(t, err, "invalid PushPop command")
}

func TestWriteArithmeticInvalidOp(t *testing.T) {
	cw := &CodeWriter{}
	err := cw.WriteArithmetic("bogusop")
	assert.ErrorContains(t, err, "invalid arithmetic operation")
}

func TestCloseFlushError(t *testing.T) {
	f, err := os.CreateTemp("", "x*.asm")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	w := bufio.NewWriterSize(failWriter{}, 1)
	w.WriteString(strings.Repeat("x", 100))

	cw := &CodeWriter{file: f, writer: w, asmFile: f.Name()}
	cw.Close()
}

func TestCloseFileError(t *testing.T) {
	f, err := os.CreateTemp("", "x*.asm")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	cw := &CodeWriter{file: f, writer: bufio.NewWriter(f), asmFile: f.Name()}
	cw.Close()
}

func TestPushDirectInvalidSegment(t *testing.T) {
	cw := &CodeWriter{}
	err := cw.pushDirect("notinmap", 0)
	assert.ErrorContains(t, err, "push: base address not found")
}

func TestPopDirectInvalidSegment(t *testing.T) {
	cw := &CodeWriter{}
	err := cw.popDirect("notinmap", 0)
	assert.ErrorContains(t, err, "pop: base address not found")
}
