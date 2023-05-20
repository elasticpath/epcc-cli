package completion

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	"regexp"
	"strings"
)

var doubleQuotedRegexp = regexp.MustCompile(`\s*"([^\\"]|(\\"))+"\s*`)

var singleQuotedRegexp = regexp.MustCompile(`\s*'([^\\']|(\\'))+'\s*`)

var doubleQuotedOpenRegexp = regexp.MustCompile(`\s*"([^\\"]|(\\"))+\s*$`)

var singleQuotedOpenRegexp = regexp.MustCompile(`\s*'([^\\']|(\\'))+\s*$`)

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
		if t.tokenType.typeCode == binary_op || t.tokenType.typeCode == vararg_op {
			lastOperatorToken = &tt
			commasInOperator = 0
		}

		if t.tokenType.typeCode == comma {
			commasInOperator++
		}
	}

	switch lastElement.currentState {
	case lexerState_singleQuote:
		switch lastElement.tokenType.typeCode {
		case single_quote_str_contents:
			if lastOperatorToken.tokenType.typeCode == binary_op && commasInOperator >= 1 {
				return []string{toComplete + `')`}
			} else {
				return []string{toComplete + `',`, toComplete + `')`}
			}
		case single_quote_char:
			if lastOperatorToken.tokenType.typeCode == binary_op && commasInOperator >= 1 {
				return []string{toComplete + ")"}
			} else {
				return []string{toComplete + ",", toComplete + ")"}
			}

		}
		//

	case lexerState_doubleQuote:
		switch lastElement.tokenType.typeCode {
		case double_quote_str_contents:
			if lastOperatorToken.tokenType.typeCode == binary_op && commasInOperator >= 1 {
				return []string{toComplete + `")`}
			} else {
				return []string{toComplete + `",`, toComplete + `")`}
			}
		case double_quote_char:
			if lastOperatorToken.tokenType.typeCode == binary_op && commasInOperator >= 1 {
				return []string{toComplete + ")"}
			} else {
				return []string{toComplete + ",", toComplete + ")"}
			}

		}

	case lexerState_regular:
		switch lastElement.tokenType.typeCode {
		case binary_op, vararg_op:
			for _, attr := range attributeNames {
				completions = append(completions, lastElement.entireTextMatchFromStart+attr+",")
			}
		case chain:
			for _, op := range ops {
				completions = append(completions, lastElement.entireTextMatchFromStart+op)
			}
		case raw_literal:
			if len(res) >= 2 {
				secondLastElement := res[len(res)-2]
				for _, op := range ops {
					completions = append(completions, secondLastElement.entireTextMatchFromStart+op)
				}
			} else {
				completions = append(completions, ops...)
			}

		}

	case lexerState_filterOp:
		// There must be two elements in the list (since the current state is filterOp, one state must have transitioned us)
		secondLastElement := res[len(res)-2]
		switch lastElement.tokenType.typeCode {
		case raw_literal:
			switch secondLastElement.tokenType.typeCode {
			case binary_op, vararg_op:
				// Previous element is a operator, so let's assume a field.
				for _, attr := range attributeNames {
					completions = append(completions, secondLastElement.entireTextMatchFromStart+attr+",")
				}
			case comma:
				if lastOperatorToken.tokenType.typeCode == binary_op {
					return []string{toComplete + `)`}
				} else {
					return []string{toComplete + `,`, toComplete + `)`}
				}

			}
		case right_parenthesis:
			completions = append(completions, lastElement.entireTextMatchFromStart+":")
		}

	}

	return completions

}

type tokenCode uint16

const (
	chain tokenCode = 1 << iota
	binary_op
	vararg_op
	right_parenthesis
	comma
	single_quote_char
	single_quote_str_contents
	double_quote_char

	double_quote_str_contents

	raw_literal
)

type lexerState uint8

const (
	lexerState_regular lexerState = 1 << iota
	lexerState_filterOp

	lexerState_singleQuote
	lexerState_doubleQuote
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
		validStates:  map[lexerState]lexerState{lexerState_regular: lexerState_regular},
	},
	{
		typeCode:     binary_op,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*(eq|like|gt|ge|lt|le)\s*[(]`)},
		validStates:  map[lexerState]lexerState{lexerState_regular: lexerState_filterOp},
	},
	{
		typeCode:     vararg_op,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*(in)\s*[(]`)},
		validStates:  map[lexerState]lexerState{lexerState_regular: lexerState_filterOp},
	},
	//{
	//	typeCode:     left_parenthesis,
	//	regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*\(`)},
	//	validStates:  map[lexerState]lexerState{lexerState_regular: lexerState_regular},
	//},
	{
		typeCode:     right_parenthesis,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*\)`)},
		validStates:  map[lexerState]lexerState{lexerState_filterOp: lexerState_regular},
	},
	{
		typeCode:     comma,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*,`)},
		validStates:  map[lexerState]lexerState{lexerState_filterOp: lexerState_filterOp},
	},
	{
		typeCode:     single_quote_char,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*'`)},
		validStates:  map[lexerState]lexerState{lexerState_filterOp: lexerState_singleQuote, lexerState_singleQuote: lexerState_filterOp},
	},
	{
		typeCode:     double_quote_char,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*"`)},
		validStates:  map[lexerState]lexerState{lexerState_filterOp: lexerState_doubleQuote, lexerState_doubleQuote: lexerState_filterOp},
	},
	{
		typeCode:     single_quote_str_contents,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*([^\\']|(\\'))+`)},
		validStates:  map[lexerState]lexerState{lexerState_singleQuote: lexerState_singleQuote},
	},
	{
		typeCode:     double_quote_str_contents,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*([^\\"]|(\\"))+`)},
		validStates:  map[lexerState]lexerState{lexerState_doubleQuote: lexerState_doubleQuote},
	},
	{
		typeCode:     raw_literal,
		regexMatches: []*regexp.Regexp{regexp.MustCompile(`^\s*[a-zA-Z0-9@$_*.{}| +:/-]+`)},
		// A raw literal is what we parse incomplete functions to be
		validStates: map[lexerState]lexerState{lexerState_regular: lexerState_regular, lexerState_filterOp: lexerState_filterOp},
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
	var currentState = lexerState_regular

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
