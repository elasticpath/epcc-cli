package completion

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	"regexp"
	"strings"
)

func GetFilterCompletion(toComplete string, r resources.Resource) []string {

	res, err := lex(toComplete)

	ops := []string{"eq(", "ge(", "gt(", "in(", "le(", "like(", "lt("}
	if err != nil || len(res) == 0 {
		return ops
	}

	fmt.Sprintf("%v", res)

	lastElement := res[len(res)-1]

	attributeNames := make([]string, 0, len(r.Attributes))

	for _, attr := range r.Attributes {
		attributeNames = append(attributeNames, attr.Key)
	}

	attributeNames = append(attributeNames, "updated_at", "created_at", "id")

	completions := make([]string, 0)

	var lastOperatorToken *lexedToken = nil
	var commasInOperator = 0
	for _, t := range res {
		tt := t
		if t.tokenType.typeCode == binaryOp || t.tokenType.typeCode == varargOp {
			lastOperatorToken = &tt
			commasInOperator = 0
		}

		if t.tokenType.typeCode == comma {
			commasInOperator++
		}
	}

	switch lastElement.currentState {
	case lexerStateInSingleQuote:
		switch lastElement.tokenType.typeCode {
		case singleQuoteStrContents:
			if lastOperatorToken.tokenType.typeCode == binaryOp && commasInOperator >= 1 {
				return []string{toComplete + `')`}
			} else {
				return []string{toComplete + `',`, toComplete + `')`}
			}
		case singleQuoteChar:
			if lastOperatorToken.tokenType.typeCode == binaryOp && commasInOperator >= 1 {
				return []string{toComplete + ")"}
			} else {
				return []string{toComplete + ",", toComplete + ")"}
			}

		}
		//

	case lexerStateInDoubleQuote:
		switch lastElement.tokenType.typeCode {
		case doubleQuoteStrContents:
			if lastOperatorToken.tokenType.typeCode == binaryOp && commasInOperator >= 1 {
				return []string{toComplete + `")`}
			} else {
				return []string{toComplete + `",`, toComplete + `")`}
			}
		case doubleQuoteChar:
			if lastOperatorToken.tokenType.typeCode == binaryOp && commasInOperator >= 1 {
				return []string{toComplete + ")"}
			} else {
				return []string{toComplete + ",", toComplete + ")"}
			}

		}

	case lexerStateRegular:
		switch lastElement.tokenType.typeCode {
		case binaryOp, varargOp:
			for _, attr := range attributeNames {
				completions = append(completions, lastElement.entireTextMatchFromStart+attr+",")
			}
		case chain:
			for _, op := range ops {
				completions = append(completions, lastElement.entireTextMatchFromStart+op)
			}
		case rawLiteral:
			if len(res) >= 2 {
				secondLastElement := res[len(res)-2]
				for _, op := range ops {
					completions = append(completions, secondLastElement.entireTextMatchFromStart+op)
				}
			} else {
				completions = append(completions, ops...)
			}

		}

	case lexerStateInFilterOperator:
		// There must be two elements in the list (since the current state is filterOp, one state must have transitioned us)
		secondLastElement := res[len(res)-2]
		switch lastElement.tokenType.typeCode {
		case rawLiteral:
			switch secondLastElement.tokenType.typeCode {
			case binaryOp, varargOp:
				// Previous element is a operator, so let's assume a field.
				for _, attr := range attributeNames {
					completions = append(completions, secondLastElement.entireTextMatchFromStart+attr+",")
				}
			case comma:
				if lastOperatorToken.tokenType.typeCode == binaryOp {
					return []string{toComplete + `)`}
				} else {
					return []string{toComplete + `,`, toComplete + `)`}
				}

			}
		case rightParenthesis:
			completions = append(completions, lastElement.entireTextMatchFromStart+":")
		}

	}

	return completions

}

type tokenCode uint16

const (
	chain tokenCode = 1 << iota
	binaryOp
	varargOp
	rightParenthesis
	comma
	singleQuoteChar
	singleQuoteStrContents
	doubleQuoteChar
	doubleQuoteStrContents
	rawLiteral
)

type lexerState uint8

const (
	lexerStateRegular lexerState = 1 << iota
	lexerStateInFilterOperator

	lexerStateInSingleQuote
	lexerStateInDoubleQuote
)

type tokenType struct {
	typeCode     tokenCode
	regexMatches []*regexp.Regexp
	validStates  map[lexerState]lexerState
}

var tokenTypes = []tokenType{
	{
		typeCode:     chain,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*:`)},
		validStates:  map[lexerState]lexerState{lexerStateRegular: lexerStateRegular},
	},
	{
		typeCode:     binaryOp,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*(eq|like|gt|ge|lt|le)\s*[(]`)},
		validStates:  map[lexerState]lexerState{lexerStateRegular: lexerStateInFilterOperator},
	},
	{
		typeCode:     varargOp,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*(in)\s*[(]`)},
		validStates:  map[lexerState]lexerState{lexerStateRegular: lexerStateInFilterOperator},
	},
	{
		typeCode:     rightParenthesis,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*\)`)},
		validStates:  map[lexerState]lexerState{lexerStateInFilterOperator: lexerStateRegular},
	},
	{
		typeCode:     comma,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*,`)},
		validStates:  map[lexerState]lexerState{lexerStateInFilterOperator: lexerStateInFilterOperator},
	},
	{
		typeCode:     singleQuoteChar,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*'`)},
		validStates:  map[lexerState]lexerState{lexerStateInFilterOperator: lexerStateInSingleQuote, lexerStateInSingleQuote: lexerStateInFilterOperator},
	},
	{
		typeCode:     doubleQuoteChar,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*"`)},
		validStates:  map[lexerState]lexerState{lexerStateInFilterOperator: lexerStateInDoubleQuote, lexerStateInDoubleQuote: lexerStateInFilterOperator},
	},
	{
		typeCode:     singleQuoteStrContents,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*([^\\']|(\\'))+`)},
		validStates:  map[lexerState]lexerState{lexerStateInSingleQuote: lexerStateInSingleQuote},
	},
	{
		typeCode:     doubleQuoteStrContents,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*([^\\"]|(\\"))+`)},
		validStates:  map[lexerState]lexerState{lexerStateInDoubleQuote: lexerStateInDoubleQuote},
	},
	{
		typeCode:     rawLiteral,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*[a-zA-Z0-9@$_*.{}| +:/-]+`)},
		// A raw literal is what we parse incomplete functions to be
		validStates: map[lexerState]lexerState{lexerStateRegular: lexerStateRegular, lexerStateInFilterOperator: lexerStateInFilterOperator},
	},
}

type lexedToken struct {
	tokenType                *tokenType
	text                     string
	entireTextMatchFromStart string

	currentState lexerState
	nextState    lexerState
}

func lex(input string) ([]lexedToken, error) {
	var currentState = lexerStateRegular

	var remainingString = strings.Trim(input, " ")

	lexedTokens := make([]lexedToken, 0)
	var entireMatch = ""

next:
	for len(remainingString) > 0 {
		for _, tokenType := range tokenTypes {
			for lexerState, nextState := range tokenType.validStates {
				if lexerState == currentState {
					for _, r := range tokenType.regexMatches {
						if match := r.FindString(remainingString); len(match) > 0 {
							remainingString = remainingString[len(match):]

							entireMatch = entireMatch + match

							lexedTokens = append(lexedTokens, lexedToken{
								tokenType:                &tokenType,
								text:                     match,
								currentState:             currentState,
								nextState:                nextState,
								entireTextMatchFromStart: entireMatch,
							})

							currentState = nextState
							continue next
						}
					}
				}
			}

		}
		return nil, fmt.Errorf("could not lex [%s], remaining [%s]", input, remainingString)
	}

	return lexedTokens, nil

}
