package codewriter

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctjnkns/nand2tetris/08/vmtranslator/parser"
)

const (
	asmExtension = ".asm"
)

type CodeWriter struct {
	file         *os.File
	writer       *bufio.Writer
	asmFile      string
	staticPrefix string
	labelNum     int
}

func NewCodeWriter(argument string, initSPManually bool) (*CodeWriter, error) {
	fileName := filepath.Base(argument)
	filePath := filepath.Dir(argument)

	trimmedFileName := strings.TrimSuffix(fileName, parser.VMExtension)
	asmFileName := trimmedFileName + asmExtension

	asmFile := filepath.Join(filePath, asmFileName)
	f, err := os.Create(asmFile)
	if err != nil {
		return nil, fmt.Errorf("unable to create .asm file: %v", err)
	}

	w := bufio.NewWriter(f)

	cw := &CodeWriter{
		file:         f,
		writer:       w,
		asmFile:      asmFile,
		staticPrefix: trimmedFileName,
	}

	if initSPManually {
		err := cw.initSP()
		if err != nil {
			return nil, err
		}
	}

	return cw, nil
}

func (cw *CodeWriter) initSP() error {
	lines := []string{
		"// init SP",
		"@256",
		"D=A",
		"@SP",
		"M=D",
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) writeLines(lines []string) error {
	for _, line := range lines {
		_, err := cw.writer.WriteString(line)
		if err != nil {
			return err
		}
		err = cw.writer.WriteByte('\n')
		if err != nil {
			return err
		}
	}

	return nil
}

func (cw *CodeWriter) Close() {
	err := cw.writer.Flush()
	if err != nil {
		log.Printf("failed to flush asm writer: %s", err)
	} else {
		log.Printf("assembly code written to: %s", cw.asmFile)
	}

	err = cw.file.Close()
	if err != nil {
		log.Printf("failed to close asm file: %v", err)
	}
}
