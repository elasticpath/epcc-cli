package completion

import (
	"fmt"
	"regexp"

	"github.com/quasilyte/regex/syntax"
)

func NewRegexCompletionTree() *RegexTree {
	return &RegexTree{}
}

type RegexTree struct {
	rootNode *regexNode
}

type regexNode struct {
	expr     *syntax.Expr
	depth    int
	children map[string]*regexNode
	// Full regex prefix
	prefix      string
	prefixRegex *regexp.Regexp
	parent      *regexNode

	matchesSet map[string]bool
}

var arrayIndexRegex = regexp.MustCompile("\\[0-9+]")

func (t *RegexTree) AddRegex(regex string) error {

	if regex[0] != '^' || regex[len(regex)-1] != '$' {
		return fmt.Errorf("regex needs to start with a ^ and end with a $, not: %v", regex)
	}

	p := syntax.NewParser(&syntax.ParserOptions{})
	parse, err := p.Parse(regex)

	if err != nil {
		return err
	}

	if parse.Expr.Op != syntax.OpConcat {
		return fmt.Errorf("unknown root operation: %v for regex: %v", parse.Expr.Op, parse.Pattern)
	}

	if len(parse.Expr.Args) == 0 {
		return fmt.Errorf("regex had no arguments, this was unexpected: %v", parse.Pattern)
	}

	if t.rootNode == nil {

		regex, err := regexp.Compile(parse.Expr.Args[0].Value)

		if err != nil {
			return fmt.Errorf("couldn't compile regex: %w", err)
		}

		t.rootNode = &regexNode{
			expr:        &parse.Expr.Args[0],
			depth:       0,
			children:    map[string]*regexNode{},
			prefix:      parse.Expr.Args[0].Value,
			prefixRegex: regex,
			matchesSet:  map[string]bool{},
		}
	}

	return t.rootNode.AddRegex(parse.Expr.Args, 0)
}

func (t *RegexTree) GetCompletionOptions() ([]string, error) {

	if t.rootNode == nil {
		return []string{}, nil
	}

	return t.rootNode.GetCompletionOptions()
}

func (t *RegexTree) AddExistingValue(v string) error {
	if t.rootNode == nil {
		return fmt.Errorf("tree not initialized")
	}

	newVal := arrayIndexRegex.ReplaceAllString(v, "[n]")

	return t.rootNode.AddExistingValue(newVal)
}

func (n *regexNode) AddRegex(parseTree []syntax.Expr, cdepth int) error {
	if n.expr.Value != parseTree[cdepth].Value {
		return fmt.Errorf("requested to add an expression of type %v at depth %d that collides with our current type %v", parseTree[cdepth].Value, cdepth, n.expr.Value)
	}

	if len(parseTree) == cdepth+1 {
		// We are done
		return nil
	} else if len(parseTree) < cdepth {
		panic(fmt.Sprintf("The set of expressions %d is less than our current depth %d , this is likely a bug", len(parseTree), cdepth))
	}

	nextExpr := parseTree[cdepth+1]

	nextNode := n.children[nextExpr.Value]

	if nextNode == nil {

		r := n.prefix + nextExpr.Value
		nextRegex, err := regexp.Compile(r)

		if n.expr.Op == syntax.OpDot {
			// regex . means wildcard, but in json it's the path separator.
			// I think 99 times out of a 100 this is a type
			return fmt.Errorf("regex can't use a dot, this should probably be escaped: %v", r)
		}

		if err != nil {
			return fmt.Errorf("couldn't compile regex: %w", err)
		}

		nextNode = &regexNode{
			expr:        &nextExpr,
			depth:       cdepth + 1,
			children:    map[string]*regexNode{},
			prefix:      r,
			prefixRegex: nextRegex,
			parent:      n,
			matchesSet:  map[string]bool{},
		}

		n.children[nextExpr.Value] = nextNode
	}

	return nextNode.AddRegex(parseTree, cdepth+1)
}

func (n *regexNode) GetCompletionOptions() ([]string, error) {
	completionOptions := []string{}

	for _, v := range n.children {
		childOptions, err := v.GetCompletionOptions()
		if err != nil {
			return nil, err
		}
		completionOptions = append(completionOptions, childOptions...)
	}

	switch n.expr.Op {
	case syntax.OpCaret:
	case syntax.OpDollar:
		return []string{""}, nil
	case syntax.OpLiteral:
		for k, v := range completionOptions {
			completionOptions[k] = n.expr.Value + v
		}
	case syntax.OpCapture:
		newCompletionOptions := make([]string, 0, len(completionOptions)*len(n.matchesSet))
		for _, cV := range completionOptions {
			for nV := range n.matchesSet {
				newCompletionOptions = append(newCompletionOptions, nV+cV)
			}
		}

		if len(newCompletionOptions) == 0 {
			// In this case we are a capture group, but have no examples and so can't continue
			return []string{""}, nil
		} else {
			return newCompletionOptions, nil
		}

	case syntax.OpChar:
		for k, v := range completionOptions {
			completionOptions[k] = n.expr.Value + v
		}

	case syntax.OpEscapeMeta:

		if n.expr.Value == "\\." {
			for k, v := range completionOptions {
				completionOptions[k] = "." + v
			}
		} else if n.expr.Value == "\\[" {
			for k, v := range completionOptions {
				completionOptions[k] = "[" + v
			}
		} else if n.expr.Value == "\\]" {
			for k, v := range completionOptions {
				completionOptions[k] = "]" + v
			}
		} else {
			fmt.Errorf("unable to handle regex node type %v with value %v", n.expr.Op, n.expr.Value)
		}
	default:
		return nil, fmt.Errorf("unable to handle regex node type %v", n.expr.Op)
	}

	return completionOptions, nil
}

func (n *regexNode) AddExistingValue(v string) error {

	match := n.prefixRegex.FindString(v)

	if match == "" && n.expr.Op != syntax.OpCaret {
		// We don't match this value
		return nil
	}

	parentMatch := ""

	if n.parent != nil {
		parentMatch = n.parent.prefixRegex.FindString(v)
	}

	currentMatch := match[len(parentMatch):]

	n.matchesSet[currentMatch] = true

	var cError error = nil

	for _, c := range n.children {
		cError = c.AddExistingValue(v)

		if cError != nil {
			return cError
		}
	}

	return nil
}
