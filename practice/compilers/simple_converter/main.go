package main

import (
	"fmt"
	"github.com/er1c-zh/go-now/log"
)

func main()  {
	defer log.Flush()

	tokenList := LexicalAnalyze("1+2+3-3")

	log.Debug("%v", tokenList)
}

func LexicalAnalyze(s string) []Token {
	tokenList := make([]Token, 0, len(s))
	for i, cur := range s {
		tokenList = append(tokenList, Token{
			Origin: s[i:i+1],
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
	Type TokenType
}

func (t Token) String() string {
	return fmt.Sprintf("[%d]%s", t.Type, t.Origin)
}

type Node struct {

}
