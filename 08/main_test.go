package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"

	"github.com/ctjnkns/nand2tetris/08/vmtranslator/codewriter"
	"github.com/ctjnkns/nand2tetris/08/vmtranslator/parser"
	"github.com/stretchr/testify/assert"
)

func TestTranslateScannerError(t *testing.T) {
	r := iotest.ErrReader(errors.New("scan error"))
	p := &parser.Parser{
		Scanner:     bufio.NewScanner(r),
		NextCommand: "add",
		FileName:    "test",
	}

	cw, err := codewriter.NewCodeWriter(filepath.Join(t.TempDir(), "x.asm"), false)
	if err != nil {
		t.Fatal(err)
	}
	defer cw.Close()

	vt := &VMTranslator{Parsers: []*parser.Parser{p}, CodeWriter: cw}
	err = vt.translate()
	assert.EqualError(t, err, "scan error on exit: scan error")
}

func TestMain(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
	}{
		{
			name:      "SimpleAdd",
			inputFile: "./testdata/SimpleAdd.vm",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = []string{"main", test.inputFile}
			main()
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		argument    []string // path to .vm or dir
		outputFile  string   // output file is the location the translated asm should be written to by the program when exectued
		expectedASM string   // path to the asm golden file to compare to
		wantErr     error
	}{
		{
			name:        "SimpleAdd",
			argument:    []string{"./testdata/SimpleAdd.vm"},
			outputFile:  "./testdata/SimpleAdd.asm",
			expectedASM: "./testdata/golden/SimpleAdd.asm",
		},
		{
			name:        "StackTest",
			argument:    []string{"./testdata/StackTest.vm"},
			outputFile:  "./testdata/StackTest.asm",
			expectedASM: "./testdata/golden/StackTest.asm",
		},
		{
			name:        "BasicTest",
			argument:    []string{"./testdata/BasicTest.vm"},
			outputFile:  "./testdata/BasicTest.asm",
			expectedASM: "./testdata/golden/BasicTest.asm",
		},
		{
			name:        "PointerTest",
			argument:    []string{"./testdata/PointerTest.vm"},
			outputFile:  "./testdata/PointerTest.asm",
			expectedASM: "./testdata/golden/PointerTest.asm",
		},
		{
			name:        "StaticTest",
			argument:    []string{"./testdata/StaticTest.vm"},
			outputFile:  "./testdata/StaticTest.asm",
			expectedASM: "./testdata/golden/StaticTest.asm",
		},
		{
			name:        "BasicLoop",
			argument:    []string{"./testdata/BasicLoop.vm"},
			outputFile:  "./testdata/BasicLoop.asm",
			expectedASM: "./testdata/golden/BasicLoop.asm",
		},
		{
			name:        "FibonacciSeries",
			argument:    []string{"./testdata/FibonacciSeries.vm"},
			outputFile:  "./testdata/FibonacciSeries.asm",
			expectedASM: "./testdata/golden/FibonacciSeries.asm",
		},
		{
			name:        "SimpleFunction",
			argument:    []string{"./testdata/SimpleFunction.vm"},
			outputFile:  "./testdata/SimpleFunction.asm",
			expectedASM: "./testdata/golden/SimpleFunction.asm",
		},
		{
			name:        "NestedCall",
			argument:    []string{"./testdata/NestedCall"},
			outputFile:  "./testdata/NestedCall/NestedCall.asm",
			expectedASM: "./testdata/golden/NestedCall.asm",
		},
		{
			name:        "FibonacciElement",
			argument:    []string{"./testdata/FibonacciElement"},
			outputFile:  "./testdata/FibonacciElement/FibonacciElement.asm",
			expectedASM: "./testdata/golden/FibonacciElement.asm",
		},
		{
			name:        "StaticsTest",
			argument:    []string{"./testdata/StaticsTest"},
			outputFile:  "./testdata/StaticsTest/StaticsTest.asm",
			expectedASM: "./testdata/golden/StaticsTest.asm",
		},
		{
			name:     "no input arg",
			argument: []string{}, // becomes len(args)=1 after prepend
			wantErr:  errors.New("failed to initialize vm translator: must provide .vm file as an argument"),
		},
		{
			name:     "non-existent path",
			argument: []string{"./testdata/DoesNotExist.vm"},
			wantErr:  errors.New("failed to initialize vm translator: unable to open files: stat ./testdata/DoesNotExist.vm: no such file or directory"),
		},
		{
			name:     "wrong extension",
			argument: []string{"./testdata/NotVM.txt"},
			wantErr:  errors.New("failed to initialize vm translator: file extension must be .vm: ./testdata/NotVM.txt"),
		},
		{
			name:     "empty directory",
			argument: []string{"./testdata/EmptyDir"},
			wantErr:  errors.New("failed to initialize vm translator: unable to open files: no vm files in directory"),
		},
		{
			name:     "bad glob pattern",
			argument: []string{"./testdata/[bad"},
			wantErr:  errors.New("failed to initialize vm translator: unable to open files: syntax error in pattern"),
		},
		{
			name:     "asm path collision",
			argument: []string{"./testdata/AsmCollide"},
			wantErr:  errors.New("failed to initialize vm translator: unable to create .asm file: open testdata/AsmCollide/AsmCollide.asm: is a directory"),
		},
		{
			name:       "invalid arithmetic op",
			argument:   []string{"./testdata/InvalidArith.vm"},
			outputFile: "./testdata/InvalidArith.asm",
			wantErr:    errors.New("unknown command type: bogusop"),
		},
		{
			name:     "invalid segment",
			argument: []string{"./testdata/InvalidSegment.vm"},
			wantErr:  errors.New("push: virtual segment not found in map: foo"),
		},
		{
			name:     "invalid pointer index",
			argument: []string{"./testdata/InvalidPointer.vm"},
			wantErr:  errors.New("received unsupported pointer index: 5"),
		},
		{
			name:     "missing push index",
			argument: []string{"./testdata/InvalidPushPopArg.vm"},
			wantErr:  errors.New("index must be provided for pop and push commands"),
		},
		{
			name:     "non-int index",
			argument: []string{"./testdata/BadIndex.vm"},
			wantErr:  errors.New("index must be an int, received: abc"),
		},
		{
			name:     "empty label arg",
			argument: []string{"./testdata/EmptyLabel.vm"},
			wantErr:  errors.New("arg 1: invalid vm command: label"),
		},
		{
			name:     "invalid call arg",
			argument: []string{"./testdata/InvalidCallArg.vm"},
			wantErr:  errors.New("index must be an int, received: abc"),
		},
		{
			name:     "invalid function arg",
			argument: []string{"./testdata/InvalidFuncArg.vm"},
			wantErr:  errors.New("index must be an int, received: abc"),
		},
		{
			name:     "pop invalid segment",
			argument: []string{"./testdata/InvalidPopSegment.vm"},
			wantErr:  errors.New("pop: virtual segment not found in map: foo"),
		},
		{
			name:     "pop invalid pointer",
			argument: []string{"./testdata/InvalidPopPointer.vm"},
			wantErr:  errors.New("received unsupported pointer index: 5"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := append([]string{"vmtranslator"}, test.argument...)
			err := run(args)
			if test.wantErr != nil {
				assert.EqualError(t, err, test.wantErr.Error())
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				want, err := os.ReadFile(test.expectedASM)
				if err != nil {
					t.Fatalf("failed to open expected asm file: %v", err)
				}

				got, err := os.ReadFile(test.outputFile)
				if err != nil {
					t.Fatalf("failed to open output asm file: %v", err)
				}

				assert.Equal(t, want, got, fmt.Sprintf("got:\n%s\nwant:\n%s\n", string(want), string(got)))

			}
		})
	}
}
