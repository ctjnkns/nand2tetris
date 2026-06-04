package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/compilationengine"
	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/tokenizer"
)

const (
	vmExtension = ".vm"
)

type Analyzer struct {
	tokenWriters []*tokenWriter
	tokenizeMode bool
}

type tokenWriter struct {
	tokenizer         *tokenizer.Tokenizer
	compilationEngine compilationengine.Compiler
	vmFile            *os.File
	writer            *bufio.Writer
}

func NewAnalyzer(args []string, tokenize bool) (*Analyzer, error) {
	if len(args) < 1 {
		return nil, errors.New("must provide .jack file or folder as an argument")
	}

	argument := args[0]

	info, err := os.Stat(argument)
	if err != nil {
		return nil, err
	}

	clean := filepath.Clean(argument)

	var jackFileNames []string
	if info.IsDir() {
		globPath := filepath.Join(clean, fmt.Sprintf("*%s", tokenizer.JackExtension))
		globbedFileNames, err := filepath.Glob(globPath)
		if err != nil {
			return nil, err
		}

		if len(globbedFileNames) == 0 {
			return nil, errors.New("no jack files in directory")
		}

		jackFileNames = globbedFileNames

	} else {
		jackFileNames = []string{clean}
	}

	a := &Analyzer{tokenizeMode: tokenize}

	for _, jackFile := range jackFileNames {
		if err := a.addFile(jackFile); err != nil {
			a.Close()
			return nil, err
		}
	}

	return a, nil
}

func (a *Analyzer) addFile(jackFile string) error {
	t, err := tokenizer.NewTokenizer(jackFile)
	if err != nil {
		return err
	}

	trimmedFilename := strings.TrimSuffix(jackFile, tokenizer.JackExtension)

	vmFileName := trimmedFilename + vmExtension

	vmFile, err := os.Create(vmFileName)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(vmFile)

	if !a.tokenizeMode {
		if _, err := w.WriteString(fmt.Sprintf("// Compiled %s:\n", filepath.Base(jackFile))); err != nil {
			return err
		}
	}

	ce := compilationengine.NewCompilationEngine(t, w)

	tw := &tokenWriter{
		tokenizer:         t,
		compilationEngine: ce,
		vmFile:            vmFile,
		writer:            w,
	}

	a.tokenWriters = append(a.tokenWriters, tw)

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("jackanalyzer", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [-tokenize] <file.jack|dir>\n", fs.Name())
		fs.PrintDefaults()
	}
	tokenize := fs.Bool("tokenize", false, "emit T.xml token file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	a, err := NewAnalyzer(fs.Args(), *tokenize)
	if err != nil {
		return err
	}
	defer a.Close()

	if *tokenize {
		return a.tokenize()
	}

	return a.compile()
}

func (a *Analyzer) tokenize() error {
	for _, tokenWriter := range a.tokenWriters {
		tokenWriter.writer.WriteString("<tokens>")
		tokenWriter.writer.WriteRune('\n')

		for tokenWriter.tokenizer.HasMoreTokens() {
			tokenWriter.tokenizer.Advance()

			switch tokenWriter.tokenizer.TokenType() {
			case tokenizer.KEYWORD:
				tokenWriter.writer.WriteString(fmt.Sprintf("<keyword> %s </keyword>", escapeXML(tokenWriter.tokenizer.Token())))
				tokenWriter.writer.WriteRune('\n')
			case tokenizer.SYMBOL:
				tokenWriter.writer.WriteString(fmt.Sprintf("<symbol> %s </symbol>", escapeXML(tokenWriter.tokenizer.Token())))
				tokenWriter.writer.WriteRune('\n')
			case tokenizer.IDENTIFIER:
				tokenWriter.writer.WriteString(fmt.Sprintf("<identifier> %s </identifier>", escapeXML(tokenWriter.tokenizer.Token())))
				tokenWriter.writer.WriteRune('\n')
			case tokenizer.INT_CONST:
				tokenWriter.writer.WriteString(fmt.Sprintf("<integerConstant> %s </integerConstant>", tokenWriter.tokenizer.Token()))
				tokenWriter.writer.WriteRune('\n')
			case tokenizer.STRING_CONST:
				tokenWriter.writer.WriteString(fmt.Sprintf("<stringConstant> %s </stringConstant>", escapeXML(tokenWriter.tokenizer.Token())))
				tokenWriter.writer.WriteRune('\n')
			default:
				return fmt.Errorf("unknown token type: %d", tokenWriter.tokenizer.TokenType())
			}
		}

		tokenWriter.writer.WriteString("</tokens>")
		tokenWriter.writer.WriteRune('\n')
	}
	return nil
}

func (a *Analyzer) compile() error {
	for _, tokenWriter := range a.tokenWriters {

		err := tokenWriter.compilationEngine.CompileClass()
		if err != nil {
			return err
		}
	}
	return nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func (a *Analyzer) Close() {
	for _, tokenWriter := range a.tokenWriters {
		err := tokenWriter.writer.Flush()
		if err != nil {
			log.Printf("failed to flush token writer: %s", tokenWriter.vmFile.Name())
		}

		if err := tokenWriter.vmFile.Close(); err != nil {
			log.Printf("failed to close vm file: %s\n", tokenWriter.vmFile.Name())
		}
	}
}
