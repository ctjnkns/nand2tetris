package compilationengine

import (
	"fmt"

	"github.com/ctjnkns/nand2tetris/10/jackanalyzer/tokenizer"
)

func (ce *CompilationEngine) CompileClass() error {
	ce.tokenizer.Advance() // single advance, all other methods advance before returning

	if err := ce.writeLine("<class>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeKeyword("class"); err != nil { // get the first token ready
		return err
	}

	// class name is not restricted
	if err := ce.writeIdentifier(); err != nil {
		return err
	}

	// opening curly brace expected after class identifier
	if err := ce.writeSymbol("{"); err != nil {
		return err
	}

	if err := ce.compileClassVarDecs(); err != nil {
		return err
	}

	if err := ce.compileSubroutines(); err != nil {
		return err
	}

	if err := ce.writeSymbol("}"); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</class>")
}

func (ce *CompilationEngine) CompileClassVarDec() error {
	kw, err := ce.tokenizer.KeyWord()
	if err != nil {
		return err
	}

	if kw != tokenizer.FIELD && kw != tokenizer.STATIC {
		return fmt.Errorf("token should be field or static when calling CompileClassVarDec; received: %d", kw)
	}

	if err := ce.writeLine("<classVarDec>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeVarDec(); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</classVarDec>")
}

func (ce *CompilationEngine) compileClassVarDecs() error {
	for ce.tokenizer.HasMoreTokens() && ce.tokenizer.TokenType() == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		if kw != tokenizer.FIELD && kw != tokenizer.STATIC {
			return nil
		}

		if err := ce.CompileClassVarDec(); err != nil {
			return err
		}
	}

	return nil
}
