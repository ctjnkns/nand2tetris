package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

var hackRectL = `0000000000000000
1111110000010000
0000000000010111
1110001100000110
0000000000010000
1110001100001000
0100000000000000
1110110000010000
0000000000010001
1110001100001000
0000000000010001
1111110000100000
1110111010001000
0000000000010001
1111110000010000
0000000000100000
1110000010010000
0000000000010001
1110001100001000
0000000000010000
1111110010011000
0000000000001010
1110001100000001
0000000000010111
1110101010000111`

var hackRect = `0000000000000000
1111110000010000
0000000000010111
1110001100000110
0000000000010000
1110001100001000
0100000000000000
1110110000010000
0000000000010001
1110001100001000
0000000000010001
1111110000100000
1110111010001000
0000000000010001
1111110000010000
0000000000100000
1110000010010000
0000000000010001
1110001100001000
0000000000010000
1111110010011000
0000000000001010
1110001100000001
0000000000010111
1110101010000111`

func TestMain(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		outputFile string
		wantHack   string
	}{
		{
			name:       "add",
			inputFile:  "./test_files/Add.asm",
			outputFile: "./test_files/Add.hack",
			wantHack: `0000000000000010
1110110000010000
0000000000000011
1110000010010000
0000000000000000
1110001100001000`,
		},
		{
			name:       "MaxL",
			inputFile:  "./test_files/MaxL.asm",
			outputFile: "./test_files/MaxL.hack",
			wantHack: `0000000000000000
1111110000010000
0000000000000001
1111010011010000
0000000000001010
1110001100000001
0000000000000001
1111110000010000
0000000000001100
1110101010000111
0000000000000000
1111110000010000
0000000000000010
1110001100001000
0000000000001110
1110101010000111`,
		},
		{
			name:       "RectL",
			inputFile:  "./test_files/RectL.asm",
			outputFile: "./test_files/RectL.hack",
			wantHack:   hackRectL,
		},
		{
			name:       "Max",
			inputFile:  "./test_files/Max.asm",
			outputFile: "./test_files/Max.hack",
			wantHack: `0000000000000000
1111110000010000
0000000000000001
1111010011010000
0000000000001010
1110001100000001
0000000000000001
1111110000010000
0000000000001100
1110101010000111
0000000000000000
1111110000010000
0000000000000010
1110001100001000
0000000000001110
1110101010000111`,
		},
		{
			name:       "Rect",
			inputFile:  "./test_files/Rect.asm",
			outputFile: "./test_files/Rect.hack",
			wantHack:   hackRect,
		},
		{
			name:       " Coverage",
			inputFile:  "./test_files/Coverage.asm",
			outputFile: "./test_files/Coverage.hack",
			wantHack: `0000000000000001
1110110000101010
1111110000110011
1110001100111100
1110001100101101`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = []string{"main", test.inputFile}

			main()

			data, err := os.ReadFile(test.outputFile)
			if err != nil {
				t.Fatalf("failed to open output file; %v", err)
			}
			got := string(data)

			assert.Equal(t, test.wantHack, got, fmt.Sprintf("got:\n%s\nwant:\n%s\n", got, test.wantHack))
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		wantErr   error
	}{
		{
			name:      "Assembler Error",
			arguments: []string{"main.go", "./test_files/Max.hack"},
			wantErr:   errors.New("failed to initialize assembler: file extension must be .asm: ./test_files/Max.hack"),
		},
		{
			name:      "Not Enough Arguments",
			arguments: []string{"./test_files/Max.hack"},
			wantErr:   errors.New("must provide .asm file as an argument"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := run(test.arguments)

			assert.EqualError(t, err, test.wantErr.Error())
		})
	}
}

func TestNewAssembler(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
		wantErr   error
	}{
		{
			name:      "verify error",
			inputFile: "./test_files/Max.hack",
			wantErr:   errors.New("file extension must be .asm: ./test_files/Max.hack"),
		},
		{
			name:      "missing file",
			inputFile: "missing.asm",
			wantErr:   errors.New("open missing.asm: no such file or directory"),
		},
		{
			name:      "success",
			inputFile: "./test_files/Max.asm",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			a, err := NewAssembler(test.inputFile)

			if test.wantErr != nil {
				assert.EqualError(t, err, test.wantErr.Error())
			} else {
				assert.NotNil(t, a, "assembler should not be nil")
			}
		})
	}
}

func TestSecondPass(t *testing.T) {
	t.Run("Scanner Error", func(t *testing.T) {
		r := iotest.ErrReader(errors.New("scan error"))
		p := &parser{
			scanner:      bufio.NewScanner(r),
			nextCommand:  "@1",
			filePath:     t.TempDir(),
			hackFileName: "x.hack",
		}

		a := &assembler{parser: p, symbolTable: NewSymbolTable()}

		err := a.secondPass()
		assert.EqualError(t, err, "scan error on exit: scan error")
	})

	t.Run("Write Error", func(t *testing.T) {
		p := &parser{
			scanner:      bufio.NewScanner(strings.NewReader("")),
			nextCommand:  "",
			filePath:     filepath.Join(t.TempDir(), "does-not-exist"),
			hackFileName: "x.hack",
		}
		a := &assembler{parser: p, symbolTable: NewSymbolTable()}

		err := a.secondPass()
		assert.ErrorContains(t, err, "failed to write hack file")
	})
}
