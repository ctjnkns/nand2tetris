package parser

import (
	"bufio"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

func TestNewParser(t *testing.T) {
	_, err := NewParser("invalid")
	assert.EqualError(t, err, "file extension must be .vm: invalid")
}

func TestArg1(t *testing.T) {
	p := Parser{CurrentCommand: "invalid"}
	_, err := p.Arg1()
	assert.EqualError(t, err, "unknown command type: invalid")
}

func TestArg1ReturnDefault(t *testing.T) {
	p := &Parser{CurrentCommand: "return"}
	_, err := p.Arg1()
	assert.EqualError(t, err, "arg 1: unknown command type: 7")
}

func TestArg2LabelDefault(t *testing.T) {
	p := &Parser{CurrentCommand: "label LOOP"}
	_, err := p.Arg2()
	assert.ErrorContains(t, err, "unrecognized command type")
}

func TestArg2BadCommandType(t *testing.T) {
	p := &Parser{CurrentCommand: "bogusop"}
	_, err := p.Arg2()
	assert.ErrorContains(t, err, "unknown command type")
}

func TestNewParserMissingFile(t *testing.T) {
	_, err := NewParser("nonexistent.vm")
	assert.ErrorContains(t, err, "no such file")
}

func TestErrScannerError(t *testing.T) {
	r := iotest.ErrReader(errors.New("scan error"))
	p := &Parser{Scanner: bufio.NewScanner(r)}
	p.Scanner.Scan()
	assert.EqualError(t, p.Err(), "scan error")
}
