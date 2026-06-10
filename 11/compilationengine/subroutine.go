package compilationengine

import (
	"fmt"

	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/symboltable"
	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/tokenizer"
)

func (ce *CompilationEngine) CompileSubroutine() error {
	if ce.tokenizer.TokenType() != tokenizer.KEYWORD {
		return fmt.Errorf("CompileSubroutine requires keyword; got: %s", ce.tokenizer.TokenTypeString())
	}

	kw, err := ce.tokenizer.KeyWord()
	if err != nil {
		return err
	}

	if kw != tokenizer.CONSTRUCTOR && kw != tokenizer.FUNCTION && kw != tokenizer.METHOD {
		return fmt.Errorf("token should be constructor, function, or method when calling ComipleSubroutine; got: %s", ce.tokenizer.Token())
	}

	ce.subroutineTable.Reset()

	if kw == tokenizer.METHOD {
		if err := ce.subroutineTable.Define("this", ce.className, symboltable.ARG); err != nil {
			return err
		}
	}

	if err := ce.consumeKeyword(""); err != nil {
		return err
	}

	if err := ce.consumeReturnType(); err != nil {
		return err
	}

	// subroutineName
	ce.subroutineName = ce.tokenizer.Token()
	if err := ce.consumeIdentifier(); err != nil {
		return err
	}

	if err := ce.consumeSymbol("("); err != nil {
		return err
	}

	if err := ce.CompileParameterList(); err != nil {
		return err
	}

	if err := ce.consumeSymbol(")"); err != nil {
		return err
	}

	if err := ce.CompileSubroutineBody(); err != nil {
		return err
	}

	return nil
}

func (ce *CompilationEngine) CompileSubroutineBody() error {
	if err := ce.consumeSymbol("{"); err != nil {
		return err
	}

	if err := ce.compileVarDecs(); err != nil {
		return err
	}

	nLocals := ce.subroutineTable.VarCount(symboltable.VAR)

	ce.indent = 0
	if err := ce.writeLine(fmt.Sprintf("function %s.%s %d", ce.className, ce.subroutineName, nLocals)); err != nil {
		return err
	}

	ce.indent = 2

	if err := ce.CompileStatements(); err != nil {
		return err
	}

	if err := ce.consumeSymbol("}"); err != nil {
		return err
	}

	return nil
}

func (ce *CompilationEngine) CompileParameterList() error {
	if ce.tokenizer.Token() == ")" {
		return nil
	}

	for {
		typeName := ce.tokenizer.Token()
		if err := ce.consumeType(); err != nil {
			return err
		}

		name := ce.tokenizer.Token()
		if err := ce.subroutineTable.Define(name, typeName, symboltable.ARG); err != nil {
			return err
		}

		if err := ce.consumeIdentifier(); err != nil {
			return err
		}

		if ce.tokenizer.Token() != "," {
			break
		}

		if err := ce.consumeSymbol(","); err != nil {
			return err
		}
	}

	return nil
}

func (ce *CompilationEngine) compileSubroutines() error {
	for ce.tokenizer.HasMoreTokens() && ce.tokenizer.TokenType() == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		if kw != tokenizer.CONSTRUCTOR && kw != tokenizer.FUNCTION && kw != tokenizer.METHOD {
			return nil
		}

		if err := ce.CompileSubroutine(); err != nil {
			return err
		}
	}

	return nil
}

func (ce *CompilationEngine) CompileVarDec() error {
	kw, err := ce.tokenizer.KeyWord()
	if err != nil {
		return err
	}

	if kw != tokenizer.VAR {
		return fmt.Errorf("token should be var when calling CompileVarDec; received: %d", kw)
	}

	if err := ce.writeVarDec(); err != nil {
		return err
	}

	return nil
}

func (ce *CompilationEngine) compileVarDecs() error {
	for ce.tokenizer.HasMoreTokens() && ce.tokenizer.TokenType() == tokenizer.KEYWORD {

		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		if kw != tokenizer.VAR {
			return nil
		}

		if err := ce.CompileVarDec(); err != nil {
			return err
		}

	}

	return nil
}
