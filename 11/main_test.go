package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		argument []string
		outputs  map[string]string // outputFile -> expectedXML
		wantErr  error
	}{
		{
			name:     "ExpressionLessSquare Main",
			argument: []string{"-xml", "./testdata/ExpressionLessSquare/Main.jack"},
			outputs: map[string]string{
				"./testdata/ExpressionLessSquare/Main.xml": "./testdata/golden/ExpressionLessSquare/Main.xml",
			},
		},
		{
			name:     "ExpressionLessSquare Square",
			argument: []string{"-xml", "./testdata/ExpressionLessSquare/Square.jack"},
			outputs: map[string]string{
				"./testdata/ExpressionLessSquare/Square.xml": "./testdata/golden/ExpressionLessSquare/Square.xml",
			},
		},
		{
			name:     "ExpressionLessSquare SquareGame",
			argument: []string{"-xml", "./testdata/ExpressionLessSquare/SquareGame.jack"},
			outputs: map[string]string{
				"./testdata/ExpressionLessSquare/SquareGame.xml": "./testdata/golden/ExpressionLessSquare/SquareGame.xml",
			},
		},
		{
			name:     "ExpressionLessSquare Main tokenize",
			argument: []string{"-tokenize", "./testdata/ExpressionLessSquare/Main.jack"},
			outputs: map[string]string{
				"./testdata/ExpressionLessSquare/MainT.xml": "./testdata/golden/ExpressionLessSquare/MainT.xml",
			},
		},
		{
			name:     "ExpressionLessSquare Square tokenize",
			argument: []string{"-tokenize", "./testdata/ExpressionLessSquare/Square.jack"},
			outputs: map[string]string{
				"./testdata/ExpressionLessSquare/SquareT.xml": "./testdata/golden/ExpressionLessSquare/SquareT.xml",
			},
		},
		{
			name:     "ExpressionLessSquare SquareGame tokenize",
			argument: []string{"-tokenize", "./testdata/ExpressionLessSquare/SquareGame.jack"},
			outputs: map[string]string{
				"./testdata/ExpressionLessSquare/SquareGameT.xml": "./testdata/golden/ExpressionLessSquare/SquareGameT.xml",
			},
		},
		{
			name:     "Square",
			argument: []string{"-xml", "./testdata/Square"},
			outputs: map[string]string{
				"./testdata/Square/Main.xml":       "./testdata/golden/Square/Main.xml",
				"./testdata/Square/Square.xml":     "./testdata/golden/Square/Square.xml",
				"./testdata/Square/SquareGame.xml": "./testdata/golden/Square/SquareGame.xml",
			},
		},
		{
			name:     "Square tokenize",
			argument: []string{"-tokenize", "./testdata/Square"},
			outputs: map[string]string{
				"./testdata/Square/MainT.xml":       "./testdata/golden/Square/MainT.xml",
				"./testdata/Square/SquareT.xml":     "./testdata/golden/Square/SquareT.xml",
				"./testdata/Square/SquareGameT.xml": "./testdata/golden/Square/SquareGameT.xml",
			},
		},
		{
			name:     "ArrayTest",
			argument: []string{"-xml", "./testdata/ArrayTest"},
			outputs: map[string]string{
				"./testdata/ArrayTest/Main.xml": "./testdata/golden/ArrayTest/Main.xml",
			},
		},
		{
			name:     "ArrayTest tokenize",
			argument: []string{"-tokenize", "./testdata/ArrayTest"},
			outputs: map[string]string{
				"./testdata/ArrayTest/MainT.xml": "./testdata/golden/ArrayTest/MainT.xml",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := run(test.argument)
			if test.wantErr != nil {
				assert.EqualError(t, err, test.wantErr.Error())
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for out, exp := range test.outputs {
				want, err := os.ReadFile(exp)
				if err != nil {
					t.Fatalf("failed to open expected xml file %s: %v", exp, err)
				}
				want = bytes.ReplaceAll(want, []byte("\r"), nil)

				got, err := os.ReadFile(out)
				if err != nil {
					t.Fatalf("failed to open output xml file %s: %v", out, err)
				}

				assert.Equal(t, want, got, fmt.Sprintf("%s mismatch:\ngot:\n%s\nwant:\n%s\n", out, string(got), string(want)))
			}
		})
	}
}
