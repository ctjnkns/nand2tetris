package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctjnkns/nand2tetris/08/vmtranslator/codewriter"
	"github.com/ctjnkns/nand2tetris/08/vmtranslator/parser"
)

const asmExtension = ".asm"

type VMTranslator struct {
	Parsers    []*parser.Parser
	CodeWriter *codewriter.CodeWriter
}

func NewVMTranslator(args []string) (*VMTranslator, error) {
	if len(args) < 2 {
		return nil, errors.New("must provide .vm file as an argument")
	}

	argument := args[1]

	vmFileNames, bootstrap, asmPath, err := resolveArgument(argument)
	if err != nil {
		return nil, fmt.Errorf("unable to open files: %v", err)
	}

	ps := make([]*parser.Parser, 0, len(vmFileNames))
	for _, vmFileName := range vmFileNames {
		p, err := parser.NewParser(vmFileName)
		if err != nil {
			return nil, err
		}

		ps = append(ps, p)
	}

	cw, err := codewriter.NewCodeWriter(asmPath, bootstrap)
	if err != nil {
		return nil, err
	}

	return &VMTranslator{
		Parsers:    ps,
		CodeWriter: cw,
	}, nil
}

func resolveArgument(argument string) (vmFiles []string, bootstrap bool, asmPath string, err error) {
	info, err := os.Stat(argument)
	if err != nil {
		return nil, false, "", err
	}

	if info.IsDir() {
		clean := filepath.Clean(argument)
		vmFiles, err = filepath.Glob(filepath.Join(clean, "*.vm"))
		if err != nil {
			return nil, false, "", err
		}

		if len(vmFiles) == 0 {
			return nil, false, "", errors.New("no vm files in directory")
		}

		base := filepath.Base(clean)
		asmPath = filepath.Join(clean, base+".asm")
		bootstrap = true
	} else {
		vmFiles = []string{argument}
		asmPath = strings.TrimSuffix(argument, parser.VMExtension) + asmExtension
	}

	return vmFiles, bootstrap, asmPath, nil
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func (vt *VMTranslator) translate() error {
	for _, curParser := range vt.Parsers {
		vt.CodeWriter.SetFileName(curParser.FileName)
		for curParser.HasMoreLines() {
			curParser.Advance()

			cType, err := curParser.CommandType()
			if err != nil {
				return err
			}

			var arg1 string
			if cType != parser.C_RETURN {
				arg1, err = curParser.Arg1()
				if err != nil {
					return err
				}
			}

			var arg2 int
			if cType == parser.C_PUSH || cType == parser.C_POP || cType == parser.C_CALL || cType == parser.C_FUNCTION {
				arg2, err = curParser.Arg2()
				if err != nil {
					return err
				}
			}

			switch cType {
			case parser.C_ARITHMETIC:
				err = vt.CodeWriter.WriteArithmetic(arg1)
			case parser.C_LABEL:
				err = vt.CodeWriter.WriteLabel(arg1)
			case parser.C_IF:
				err = vt.CodeWriter.WriteIf(arg1)
			case parser.C_GOTO:
				err = vt.CodeWriter.WriteGoto(arg1)
			case parser.C_PUSH, parser.C_POP:
				err = vt.CodeWriter.WritePushPop(cType, arg1, arg2)
			case parser.C_CALL:
				err = vt.CodeWriter.WriteCall(arg1, arg2)
			case parser.C_FUNCTION:
				err = vt.CodeWriter.WriteFunction(arg1, arg2)
			case parser.C_RETURN:
				err = vt.CodeWriter.WriteReturn()
			}
			if err != nil {
				return err
			}
		}
		// bytes reader should never return an error but including in case the implementation changes and as a best practice
		if err := curParser.Err(); err != nil {
			return fmt.Errorf("scan error on exit: %v", err)
		}
	}

	return nil
}

func run(args []string) error {
	vt, err := NewVMTranslator(args)
	if err != nil {
		return fmt.Errorf("failed to initialize vm translator: %v", err)
	}
	defer vt.CodeWriter.Close()

	return vt.translate()
}
