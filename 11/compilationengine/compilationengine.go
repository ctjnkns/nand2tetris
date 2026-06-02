package compilationengine

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/tokenizer"
)

type Compiler interface {
	CompileClass() error
	CompileClassVarDec() error
	CompileSubroutine() error
	CompileParameterList() error
	CompileSubroutineBody() error
	CompileVarDec() error
	CompileStatements() error
	CompileLet() error
	CompileIf() error
	CompileWhile() error
	CompileDo() error
	CompileReturn() error
	CompileExpression() error
	CompileTerm() error
	CompileExpressionList() error
}

type CompilationEngine struct {
	tokenizer *tokenizer.Tokenizer
	writer    *bufio.Writer
	indent    int
}

func NewCompilationEngine(tokenizer *tokenizer.Tokenizer, writer *bufio.Writer) *CompilationEngine {
	ce := &CompilationEngine{
		tokenizer: tokenizer,
		writer:    writer,
		indent:    0,
	}

	return ce
}

func (ce *CompilationEngine) writeReturnType() error {
	if ce.tokenizer.TokenType() == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}
		if kw == tokenizer.VOID {
			return ce.writeExpectedAndAdvance(tokenizer.KEYWORD, "void")
		}
	}
	return ce.writeType()
}

func (ce *CompilationEngine) writeType() error {
	tt := ce.tokenizer.TokenType()
	if tt != tokenizer.KEYWORD && tt != tokenizer.IDENTIFIER {
		return fmt.Errorf("expected type; got: %s", ce.tokenizer.Token())
	}

	if tt == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		if kw != tokenizer.INT && kw != tokenizer.CHAR && kw != tokenizer.BOOLEAN {
			return fmt.Errorf("expected int|char|boolean; got %s", ce.tokenizer.Token())
		}
	}
	return ce.writeExpectedAndAdvance(ce.tokenizer.TokenType(), "")
}

func (ce *CompilationEngine) writeVarDec() error {
	// static|field|var
	if err := ce.writeKeyword(""); err != nil {
		return err
	}

	// int|char|boolean|className
	if err := ce.writeType(); err != nil {
		return err
	}

	// varName
	if err := ce.writeIdentifier(); err != nil {
		return err
	}

	// (, varName)*
	for ce.tokenizer.TokenType() == tokenizer.SYMBOL && ce.tokenizer.Token() == "," {
		if err := ce.writeSymbol(","); err != nil {
			return err
		}

		if err := ce.writeIdentifier(); err != nil {
			return err
		}
	}

	return ce.writeSymbol(";")
}

func (ce *CompilationEngine) writeSymbol(s string) error {
	return ce.writeExpectedAndAdvance(tokenizer.SYMBOL, s)
}

func (ce *CompilationEngine) writeKeyword(s string) error {
	return ce.writeExpectedAndAdvance(tokenizer.KEYWORD, s)
}

func (ce *CompilationEngine) writeIdentifier() error {
	return ce.writeExpectedAndAdvance(tokenizer.IDENTIFIER, "")
}

func (ce *CompilationEngine) writeExpectedAndAdvance(expectedType int, expectedToken string) error {
	if ce.tokenizer.TokenType() != expectedType || (expectedToken != "" && ce.tokenizer.Token() != expectedToken) {
		return fmt.Errorf("wrong type or token; expectedType: %d; got: %d, expectedToken: %s, got: %s", expectedType, ce.tokenizer.TokenType(), expectedToken, ce.tokenizer.Token())
	}

	return ce.writeLineAndAdvance(fmt.Sprintf("<%s> %s </%s>", ce.tokenizer.TokenTypeString(), escapeXML(ce.tokenizer.Token()), ce.tokenizer.TokenTypeString()))
}

func (ce *CompilationEngine) writeLineAndAdvance(line string) error {
	if err := ce.writeLine(line); err != nil {
		return err
	}

	return ce.checkAndAdvance()
}

func (ce *CompilationEngine) writeLine(line string) error {
	indent := strings.Repeat("  ", ce.indent)
	_, err := ce.writer.WriteString(indent + line)
	if err != nil {
		return err
	}
	_, err = ce.writer.WriteRune('\n')
	if err != nil {
		return err
	}

	return nil
}

func (ce *CompilationEngine) checkAndAdvance() error {
	if ce.tokenizer.HasMoreTokens() {
		ce.tokenizer.Advance()
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
