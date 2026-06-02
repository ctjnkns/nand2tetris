package compilationengine

import (
	"fmt"

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

	if err := ce.writeLine("<subroutineDec>"); err != nil {
		return err
	}

	ce.indent++

	// constructor|function|method
	if err := ce.writeKeyword(""); err != nil {
		return err
	}

	// (void|type)
	if err := ce.writeReturnType(); err != nil {
		return err
	}

	// subroutineName
	if err := ce.writeIdentifier(); err != nil {
		return err
	}

	if err := ce.writeSymbol("("); err != nil {
		return err
	}

	if err := ce.CompileParameterList(); err != nil {
		return err
	}

	if err := ce.writeSymbol(")"); err != nil {
		return err
	}

	if err := ce.CompileSubroutineBody(); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</subroutineDec>")
}

func (ce *CompilationEngine) CompileSubroutineBody() error {
	if err := ce.writeLine("<subroutineBody>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeSymbol("{"); err != nil {
		return err
	}

	if err := ce.compileVarDecs(); err != nil {
		return err
	}

	if err := ce.CompileStatements(); err != nil {
		return err
	}

	if err := ce.writeSymbol("}"); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</subroutineBody>")
}

func (ce *CompilationEngine) CompileParameterList() error {
	if err := ce.writeLine("<parameterList>"); err != nil {
		return err
	}

	ce.indent++

	if ce.tokenizer.Token() != ")" {
		if err := ce.writeType(); err != nil {
			return err
		}

		if err := ce.writeIdentifier(); err != nil {
			return err
		}

		for ce.tokenizer.Token() == "," {
			if err := ce.writeSymbol(","); err != nil {
				return err
			}

			if err := ce.writeType(); err != nil {
				return err
			}

			if err := ce.writeIdentifier(); err != nil {
				return err
			}
		}
	}

	ce.indent--

	return ce.writeLine("</parameterList>")
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

	if err := ce.writeLine("<varDec>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeVarDec(); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</varDec>")
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
