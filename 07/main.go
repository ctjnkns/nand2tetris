package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type VMTranslator struct {
	Parser     *Parser
	CodeWriter *CodeWriter
}

func NewVMTranslator(args []string) (*VMTranslator, error) {
	if len(args) < 2 {
		return nil, errors.New("must provide .vm file as an argument")
	}

	argument := args[1]

	var initSPManually bool
	if len(args) > 2 && strings.HasSuffix(args[2], "initSP") { // matches -initSP
		initSPManually = true
	}

	p, err := NewParser(argument)
	if err != nil {
		return nil, err
	}

	cw, err := NewCodeWriter(argument, initSPManually)
	if err != nil {
		return nil, err
	}

	return &VMTranslator{
		Parser:     p,
		CodeWriter: cw,
	}, nil
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func (vt *VMTranslator) translate() error {
	for vt.Parser.HasMoreLines() {
		vt.Parser.Advance()

		cType := vt.Parser.CommandType()
		switch cType {
		case C_ARITHMETIC:
			err := vt.CodeWriter.WriteArithmetic(vt.Parser.Arg1())
			if err != nil {
				return err
			}
		case C_PUSH:
			err := vt.CodeWriter.WritePushPop(cType, vt.Parser.Arg1(), vt.Parser.Arg2())
			if err != nil {
				return err
			}
		}
	}

	// bytes reader should never return an error but including in case the implementation changes and as a best practice
	if err := vt.Parser.Err(); err != nil {
		return fmt.Errorf("scan error on exit: %v", err)
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
