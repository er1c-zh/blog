package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/er1c-zh/go-now/log"
)

func main() {
	defer log.Flush()

	tokenList := LexicalAnalyze("1+2+333")

	log.Debug("%v", tokenList)
	result, err := Translate(tokenList)
	if err != nil {
		log.Error("Translate fail: %s", err.Error())
		return
	}
	log.Debug("%s", result)
}

func LexicalAnalyze(s string) []*Token {
	tokenList := make([]*Token, 0, len(s))
	for i, cur := range s {
		tokenList = append(tokenList, &Token{
			Origin: s[i : i+1],
			Type:   terminals[cur],
		})
	}
	return tokenList
}

var (
	terminals = map[rune]TokenType{
		'+': TPlus,
		'-': TMinus,
		'0': TDigit,
		'1': TDigit,
		'2': TDigit,
		'3': TDigit,
		'4': TDigit,
		'5': TDigit,
		'6': TDigit,
		'7': TDigit,
		'8': TDigit,
		'9': TDigit,
	}
)

type TokenType int

const (
	TNull TokenType = iota
	TDigit
	TPlus
	TMinus
)

type Token struct {
	Origin string
	Type   TokenType
}

func (t Token) OutputString() string {
	return t.Origin
}

func (t Token) String() string {
	return fmt.Sprintf("[%d]%s", t.Type, t.Origin)
}

////////////////////////////////////
// translator
////////////////////////////////////

var (
	ErrEOF = errors.New("EOF")
)

type Translator struct {
	TokenList []*Token
	IdxToRead int64
	buf       bytes.Buffer
	err       error
}

func (t *Translator) Peek() (*Token, error) {
	if int64(len(t.TokenList)) <= t.IdxToRead {
		return nil, ErrEOF
	}
	return t.TokenList[t.IdxToRead], nil
}

func (t *Translator) ReadOne() (*Token, error) {
	if int64(len(t.TokenList)) <= t.IdxToRead {
		return nil, ErrEOF
	}
	cur := t.TokenList[t.IdxToRead]
	t.IdxToRead += 1
	return cur, nil
}

func (t *Translator) Write(format string, args ...interface{}) {
	t.buf.WriteString(fmt.Sprintf(format, args...))
}

func (t *Translator) SetError(err error) {
	t.err = err
}

func (t *Translator) Error() error {
	return t.err
}

func (t *Translator) String() string {
	return t.buf.String()
}

func Translate(tokenList []*Token) (string, error) {
	t := &Translator{
		TokenList: tokenList,
		IdxToRead: 0,
	}
	Expr(t)

	if t.Error() != nil {
		return "", t.Error()
	}

	return t.String(), nil
}

// Expr expr -> digit rest
func Expr(t *Translator) {
	_, err := t.Peek()
	if err != nil {
		if err == ErrEOF {
			return
		}
		t.SetError(errors.New("expect an expr but not found"))
		return
	}
	Digit(t)
	if t.Error() != nil {
		return
	}
	Rest(t)
}

// Rest rest -> + digit rest | - digit rest | empty
func Rest(t *Translator) {
	operator, err := t.ReadOne()
	if err != nil {
		if err == ErrEOF {
			return
		}
		t.SetError(errors.New("expect an operator but not found"))
		return
	}
	switch operator.Type {
	case TMinus:
		fallthrough
	case TPlus:
		Digit(t)
		if err := t.Error(); err != nil {
			return
		}
		Rest(t)
		if err := t.Error(); err != nil {
			return
		}
		t.Write("%s ", operator.OutputString())
	default:
		MissMatch(t, operator, TMinus, TPlus)
		return
	}
}

// Digit digit -> 0 | ... | 9
func Digit(t *Translator) {
	d, err := t.ReadOne()
	if err != nil {
		t.SetError(errors.New("expect a digit but not found"))
		return
	}

	if d.Type != TDigit {
		MissMatch(t, d, TDigit)
		return
	}
	t.Write("%s ", d.OutputString())
}

func MissMatch(t *Translator, token *Token, expectType ...TokenType) {
	t.SetError(fmt.Errorf("expected %d, but get [%d]'%s'",
		expectType, token.Type, token.OutputString()))
}
