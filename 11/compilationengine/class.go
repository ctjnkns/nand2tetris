package compilationengine

import (
	"fmt"

	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/tokenizer"
)

func (ce *CompilationEngine) CompileClass() error {
	ce.tokenizer.Advance() // single advance, all other methods advance before returning

	if err := ce.consumeKeyword("class"); err != nil { // get the first token ready
		return err
	}

	// class name is not restricted
	ce.className = ce.tokenizer.Token()
	if err := ce.consumeIdentifier(); err != nil {
		return err
	}

	// opening curly brace expected after class identifier
	if err := ce.consumeSymbol("{"); err != nil {
		return err
	}

	if err := ce.compileClassVarDecs(); err != nil {
		return err
	}

	if err := ce.compileSubroutines(); err != nil {
		return err
	}

	if err := ce.consumeSymbol("}"); err != nil {
		return err
	}

	return nil
}

func (ce *CompilationEngine) CompileClassVarDec() error {
	kw, err := ce.tokenizer.KeyWord()
	if err != nil {
		return err
	}

	if kw != tokenizer.FIELD && kw != tokenizer.STATIC {
		return fmt.Errorf("token should be field or static when calling CompileClassVarDec; received: %d", kw)
	}

	if err := ce.writeVarDec(); err != nil {
		return err
	}

	return nil
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
