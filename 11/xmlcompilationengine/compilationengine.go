package xmlcompilationengine

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/symboltable"
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
	CompileExpressionList() (int, error)
}

type CompilationEngine struct {
	tokenizer       *tokenizer.Tokenizer
	writer          *bufio.Writer
	className       string
	classTable      *symboltable.SymbolTable
	subroutineTable *symboltable.SymbolTable
	indent          int
}

func NewCompilationEngine(tokenizer *tokenizer.Tokenizer, writer *bufio.Writer) *CompilationEngine {
	ce := &CompilationEngine{
		tokenizer:       tokenizer,
		writer:          writer,
		classTable:      symboltable.NewSymbolTable(),
		subroutineTable: symboltable.NewSymbolTable(),
		indent:          0,
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

		return ce.writeExpectedAndAdvance(ce.tokenizer.TokenType(), "")
	}

	name := ce.tokenizer.Token()
	return ce.writeIdentifierInfo(name, "class", "used", nil)
}

func (ce *CompilationEngine) writeVarDec() error {
	kw, err := ce.tokenizer.KeyWord()
	if err != nil {
		return err
	}

	var kind symboltable.Kind

	switch kw {
	case tokenizer.STATIC:
		kind = symboltable.STATIC
	case tokenizer.FIELD:
		kind = symboltable.FIELD
	case tokenizer.VAR:
		kind = symboltable.VAR
	default:
		return fmt.Errorf("expected static, field, or var; got %s", ce.tokenizer.Token())
	}

	// static|field|var
	if err := ce.writeKeyword(""); err != nil {
		return err
	}

	typeName := ce.tokenizer.Token()

	// int|char|boolean|className
	if err := ce.writeType(); err != nil {
		return err
	}

	name := ce.tokenizer.Token()

	var table *symboltable.SymbolTable
	if kind == symboltable.VAR {
		table = ce.subroutineTable
	} else {
		table = ce.classTable
	}

	if err := table.Define(name, typeName, kind); err != nil {
		return err
	}

	index := table.IndexOf(name)
	category := kindCategory(kind)
	// varName
	if err := ce.writeIdentifierInfo(name, category, "declared", &index); err != nil {
		return err
	}

	// (, varName)*
	for ce.tokenizer.TokenType() == tokenizer.SYMBOL && ce.tokenizer.Token() == "," {
		if err := ce.writeSymbol(","); err != nil {
			return err
		}

		name := ce.tokenizer.Token()

		if err := table.Define(name, typeName, kind); err != nil {
			return err
		}

		index := table.IndexOf(name)
		category := kindCategory(kind)

		if err := ce.writeIdentifierInfo(name, category, "declared", &index); err != nil {
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

func (ce *CompilationEngine) writeIdentifierInfo(name, category, usage string, index *int) error {
	if ce.tokenizer.TokenType() != tokenizer.IDENTIFIER {
		return fmt.Errorf("expected identifier; got %s", ce.tokenizer.Token())
	}
	if ce.tokenizer.Token() != name {
		return fmt.Errorf("expected identifier %s; got %s", name, ce.tokenizer.Token())
	}

	if err := ce.writeLine("<identifier>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeLine(fmt.Sprintf("<name> %s </name>", escapeXML(name))); err != nil {
		return err
	}

	if err := ce.writeLine(fmt.Sprintf("<category> %s </category>", category)); err != nil {
		return err
	}

	if index != nil {
		if err := ce.writeLine(fmt.Sprintf("<index> %d </index>", *index)); err != nil {
			return err
		}
	}

	if err := ce.writeLine(fmt.Sprintf("<usage> %s </usage>", usage)); err != nil {
		return err
	}

	ce.indent--

	if err := ce.writeLine("</identifier>"); err != nil {
		return err
	}

	return ce.checkAndAdvance()
}

func (ce *CompilationEngine) writeVariableUse(name string) error {
	table, ok := ce.lookup(name)
	if !ok {
		return fmt.Errorf("identifier not found: %s", name)
	}

	kind := table.KindOf(name)
	index := table.IndexOf(name)
	category := kindCategory(kind)

	return ce.writeIdentifierInfo(name, category, "used", &index)
}

func kindCategory(kind symboltable.Kind) string {
	switch kind {
	case symboltable.STATIC:
		return "static"
	case symboltable.FIELD:
		return "field"
	case symboltable.ARG:
		return "arg"
	case symboltable.VAR:
		return "var"
	default:
		return "none"
	}
}

func (ce *CompilationEngine) lookup(name string) (*symboltable.SymbolTable, bool) {
	if ce.subroutineTable.KindOf(name) != symboltable.NONE {
		return ce.subroutineTable, true
	}
	if ce.classTable.KindOf(name) != symboltable.NONE {
		return ce.classTable, true
	}
	return nil, false
}
