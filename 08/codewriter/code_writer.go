package codewriter

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type CodeWriter struct {
	file            *os.File
	writer          *bufio.Writer
	asmFile         string
	staticPrefix    string
	labelNum        int
	currentFunction string
	callCount       int
}

type CodeWriterOption func(*CodeWriter)

func WithWriter(w *bufio.Writer) CodeWriterOption {
	return func(cw *CodeWriter) {
		cw.writer = w
	}
}

func NewCodeWriter(asmPath string, bootstrap bool, opts ...CodeWriterOption) (*CodeWriter, error) {
	f, err := os.Create(asmPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create .asm file: %v", err)
	}

	w := bufio.NewWriter(f)

	cw := &CodeWriter{
		file:         f,
		writer:       w,
		asmFile:      asmPath,
		staticPrefix: strings.TrimSuffix(filepath.Base(asmPath), ".asm"),
	}

	for _, opt := range opts {
		opt(cw)
	}

	if bootstrap {
		if err := cw.WriteInit(); err != nil {
			f.Close()
			return nil, err
		}
	}

	return cw, nil
}

func (cw *CodeWriter) SetFileName(name string) {
	cw.staticPrefix = name
}

func (cw *CodeWriter) WriteInit() error {
	lines := []string{
		"// init SP",
		"@256",
		"D=A",
		"@SP",
		"M=D",
	}

	err := cw.writeLines(lines)
	if err != nil {
		return err
	}

	if err := cw.WriteCall("Sys.init", 0); err != nil {
		return err
	}

	return nil
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
