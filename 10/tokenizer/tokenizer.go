package tokenizer

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	JackExtension = ".jack"
)

const (
	KEYWORD = iota
	SYMBOL
	IDENTIFIER
	INT_CONST
	STRING_CONST
)

var symbols = map[byte]bool{
	'{': true,
	'}': true,
	'(': true,
	')': true,
	'[': true,
	']': true,
	'.': true,
	',': true,
	';': true,
	'+': true,
	'-': true,
	'*': true,
	'/': true,
	'&': true,
	'|': true,
	'<': true,
	'>': true,
	'=': true,
	'~': true,
}

const (
	CLASS = iota
	METHOD
	FUNCTION
	CONSTRUCTOR
	INT
	BOOLEAN
	CHAR
	VOID
	VAR
	STATIC
	FIELD
	LET
	DO
	IF
	ELSE
	WHILE
	RETURN
	TRUE
	FALSE
	NULL
	THIS
)

var keywords = map[string]int{
	"class":       CLASS,
	"constructor": CONSTRUCTOR,
	"function":    FUNCTION,
	"method":      METHOD,
	"field":       FIELD,
	"static":      STATIC,
	"var":         VAR,
	"int":         INT,
	"char":        CHAR,
	"boolean":     BOOLEAN,
	"void":        VOID,
	"true":        TRUE,
	"false":       FALSE,
	"null":        NULL,
	"this":        THIS,
	"let":         LET,
	"do":          DO,
	"if":          IF,
	"else":        ELSE,
	"while":       WHILE,
	"return":      RETURN,
}

var xmlTag = map[int]string{
	KEYWORD:      "keyword",
	SYMBOL:       "symbol",
	IDENTIFIER:   "identifier",
	INT_CONST:    "integerConstant",
	STRING_CONST: "stringConstant",
}

type Tokenizer struct {
	data             []byte
	currentIndex     int
	currentToken     string
	currentTokenType int
	nextToken        string
	nextTokenType    int
	lenData          int
}

func NewTokenizer(jackFile string) (*Tokenizer, error) {
	if err := verify(jackFile); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(jackFile)
	if err != nil {
		return nil, err
	}

	t := &Tokenizer{
		data:    data,
		lenData: len(data),
	}

	t.Advance()

	return t, nil
}

func verify(jackFile string) error {
	if !strings.HasSuffix(jackFile, JackExtension) {
		return fmt.Errorf("file extension must be .jack: %s", jackFile)
	}

	return nil
}

func (t *Tokenizer) Token() string {
	return t.currentToken
}

func (t *Tokenizer) TokenTypeString() string {
	return xmlTag[t.currentTokenType]
}

func (t *Tokenizer) HasMoreTokens() bool {
	return t.nextToken != ""
}

func (t *Tokenizer) Advance() {
	t.currentToken = t.nextToken
	t.currentTokenType = t.nextTokenType

	// get the next token
	for {
		for t.currentIndex < t.lenData && isWhiteSpace(t.data[t.currentIndex]) {
			t.currentIndex++
		}

		if t.currentIndex+1 < t.lenData && t.data[t.currentIndex] == '/' {
			switch t.data[t.currentIndex+1] {
			case '/':
				// skip til the end of the line
				for t.currentIndex < t.lenData && t.data[t.currentIndex] != '\n' {
					t.currentIndex++
				}
				continue
			case '*':
				// scan until */
				t.currentIndex += 2
				for t.currentIndex+1 < t.lenData && !(t.data[t.currentIndex] == '*' && t.data[t.currentIndex+1] == '/') {
					t.currentIndex++
				}
				t.currentIndex += 2 // skip past the */
				continue
			}
		}
		break
	}

	if t.currentIndex >= t.lenData {
		t.nextToken = ""
		return
	}

	start := t.currentIndex
	c := t.data[t.currentIndex]

	switch {
	case isDigit(c):
		for t.currentIndex < t.lenData && isDigit(t.data[t.currentIndex]) {
			t.currentIndex++
		}
		t.nextToken = string(t.data[start:t.currentIndex])
		t.nextTokenType = INT_CONST
	case isLetter(c):
		for t.currentIndex < t.lenData && (isLetter(t.data[t.currentIndex]) || isDigit(t.data[t.currentIndex]) || t.data[t.currentIndex] == '_') {
			t.currentIndex++
		}
		t.nextToken = string(t.data[start:t.currentIndex])
		if _, ok := keywords[t.nextToken]; ok {
			t.nextTokenType = KEYWORD
		} else {
			t.nextTokenType = IDENTIFIER
		}
	case symbols[c]:
		t.nextToken = string(c)
		t.nextTokenType = SYMBOL
		t.currentIndex++
	case c == '"':
		t.currentIndex++
		strStart := t.currentIndex
		for t.currentIndex < t.lenData && t.data[t.currentIndex] != '"' {
			t.currentIndex++
		}
		t.nextToken = string(t.data[strStart:t.currentIndex])
		t.nextTokenType = STRING_CONST
		t.currentIndex++
	}
}

func isWhiteSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func (t *Tokenizer) TokenType() int {
	return t.currentTokenType
}

func (t *Tokenizer) KeyWord() (int, error) {
	if t.currentTokenType != KEYWORD {
		return -1, fmt.Errorf("only call KeyWord when tokenType is %s; current tokenType is: %s", xmlTag[INT_CONST], xmlTag[t.currentTokenType])
	}

	if kw, ok := keywords[t.currentToken]; ok {
		return kw, nil
	} else {
		return -1, fmt.Errorf("keyword value not found in map: %s", t.currentToken)
	}
}

func (t *Tokenizer) Symbol() (rune, error) {
	if t.currentTokenType != SYMBOL {
		return 0, fmt.Errorf("only call Symbol when tokenType is %s; current tokenType is: %s", xmlTag[SYMBOL], xmlTag[t.currentTokenType])
	}

	return rune(t.currentToken[0]), nil
}

func (t *Tokenizer) Identifier() (string, error) {
	if t.currentTokenType != IDENTIFIER {
		return "", fmt.Errorf("only call Identifier when tokenType is %s; current tokenType is: %s", xmlTag[IDENTIFIER], xmlTag[t.currentTokenType])
	}

	return t.currentToken, nil
}

func (t *Tokenizer) IntVal() (int, error) {
	if t.currentTokenType != INT_CONST {
		return -1, fmt.Errorf("only call IntVal when tokenType is %s; current tokenType is: %s", xmlTag[INT_CONST], xmlTag[t.currentTokenType])
	}

	cTokenInt, err := strconv.Atoi(t.currentToken)
	if err != nil {
		return -1, err
	}
	return cTokenInt, nil
}

func (t *Tokenizer) StringVal() (string, error) {
	if t.currentTokenType != STRING_CONST {
		return "", fmt.Errorf("only call StringVal when tokenType is %s; current tokenType is: %s", xmlTag[STRING_CONST], xmlTag[t.currentTokenType])
	}

	return t.currentToken, nil
}
