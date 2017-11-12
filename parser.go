package main

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
	"strings"
)

type Template struct {
	Components []TemplateNode
}

func (t *Template) Expand(c *Context) (string, error) {
	res := ""
	for _, node := range(t.Components) {
		exp, err :=  node.Expand(c)
		if err != nil {
			return res, err
		}
		res = res + exp
	}
	return res, nil
}

func (t *Template) Variables() []string {
	exp := map[string]bool{}
	for _, n := range(t.Components) {
		for _, v := range(n.Vars()) {
			exp[v] = true
		}
	}
	vars := []string{}
	for v := range(exp) {
		vars = append(vars, v)
	}
	return vars
}

func ParseString(template string) (Template, error) {
	tmpl := Template{
		Components: []TemplateNode{},
	}
	for t := range Tokenize(template) {
		if t.Kind == TOKEN_ERROR {
			return tmpl, t.Err
		} else if t.Kind == TOKEN_STRING {
			tmpl.Components = append(tmpl.Components, StringLiteralNode{Value: t.Value})
		} else if t.Kind == TOKEN_VAR {
			tmpl.Components = append(tmpl.Components, NewExpansionNode(t.Value))
		} else {
			panic("unknown token type found during parsing")
		}
	}
	return tmpl, nil
}

func Tokenize(template string) chan Token {
	token_stream := make(chan Token)
	go RunTokenizer(template, token_stream)
	return token_stream
}

const (
	TOKEN_STRING = iota
	TOKEN_VAR = iota
	TOKEN_ERROR = iota
)

const (
	STATE_STRING = iota
	STATE_STRING_ESCAPE = iota
	STATE_NAME_FIRST = iota
	STATE_NAME_AFTER = iota
	STATE_SPECIFIER_ZERO_FILL = iota
	STATE_SPECIFIER_FIELD_WIDTH = iota
	STATE_SPECIFIER_DECIMAL = iota
	STATE_NAME_COMPLETE = iota
)

func RunTokenizer(template string, out chan Token) {
	t := ""
	state := STATE_STRING
	defer close(out)
	for _, r := range template {
		if (state == STATE_STRING) {
			if r == '{' {
				out <- StringToken(t)
				t = ""
				state = STATE_NAME_FIRST
			} else if r == '\\' {
				state = STATE_STRING_ESCAPE
			} else {
				t = t + string(r)
				state = STATE_STRING
			}
		} else if state == STATE_STRING_ESCAPE {
			if r == '{' || r == '\\' {
				t = t + string(r)
				state = STATE_STRING
			} else {
				out <- ErrorToken("unknown escape code")
				return
			}
		} else if ( state == STATE_NAME_FIRST) {
			if r == '}' {
				out <- ErrorToken("variable not defined")
				return
			} else if unicode.IsLetter(r) {
				t = t + string(r)
				state = STATE_NAME_AFTER
			} else {
				out <- ErrorToken("variable nume must start with letters")
			}
		} else if ( state == STATE_NAME_AFTER) {
			if r == '}' {
				out <- VarToken(t)
				t = ""
				state = STATE_STRING
			} else if r == ':' {
				t = t + string(r)
				state = STATE_SPECIFIER_ZERO_FILL
			} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
				t = t + string(r)
				state = STATE_NAME_AFTER
			} else {
				out <- ErrorToken("variable nume must start with letters")
			}
		} else if ( state == STATE_SPECIFIER_ZERO_FILL) {
			if r == '0' {
				t = t + string(r)
				state = STATE_SPECIFIER_FIELD_WIDTH
			} else {
				out <- ErrorToken("only zero fill allowed in specifier")
				return
			}
		} else if ( state == STATE_SPECIFIER_FIELD_WIDTH) {
			if unicode.IsDigit(r) {
				t = t + string(r)
				state = STATE_SPECIFIER_DECIMAL
			} else {
				out <- ErrorToken("only digit allowed in field width")
				return
			}
		} else if ( state == STATE_SPECIFIER_DECIMAL) {
			if r == 'd' {
				t = t + string(r)
				state = STATE_NAME_COMPLETE
			} else {
				out <- ErrorToken("d expected as field type specifier")
				return
			}
		} else if ( state == STATE_NAME_COMPLETE) {
			if r == '}' {
				out <- VarToken(t)
				t = ""
				state = STATE_STRING
			} else {
				out <- ErrorToken("d expected as field type specifier")
				return
			}
		} else {
			panic("unreachable state")
		}
	}
	if state == STATE_STRING {
		if t != "" {
			out <- StringToken(t)
		}
	} else {
		out <- ErrorToken("end of string malformed")
	}
}

type Token struct {
	Kind int
	Value string
	Err error
}

func StringToken(value string) Token {
	return Token{
		Kind: TOKEN_STRING,
		Value: value,
		Err: nil,
	}
}

func VarToken(value string) Token {
	return Token{
		Kind: TOKEN_VAR,
		Value: value,
		Err: nil,
	}
}

func ErrorToken(msg string) Token {
	return Token{
		Kind: TOKEN_ERROR,
		Value: "",
		Err: errors.New(msg),
	}
}

type TemplateNode interface {
	Expand(c *Context) (string, error)
	Vars() []string
}

type StringLiteralNode struct {
	Value string
}

func (n StringLiteralNode) Expand(c *Context) (string, error) {
	return n.Value, nil
}

func (n StringLiteralNode) Vars() []string {
	return []string{}
}

type ExpansionNode struct {
	Name string
}

func NewExpansionNode(varExpr string) TemplateNode {
	parts := strings.Split(varExpr, ":")
	if len(parts) == 1 {
		return &ExpansionNode{Name: parts[0]}
	} else if len(parts) == 2 {
		name := parts[0]
		widthSpeciferRune := []rune(parts[1])[1]
		widthSpecifer, err := strconv.Atoi(string(widthSpeciferRune))
		if err != nil {
			panic("run specifier must be a digit")
		}
		return &ZeroFillExpansionNode{Name: name, FieldWidth: widthSpecifer}
	} else {
		panic("malformed expansion node definition")
	}
}

func (n ExpansionNode) Expand(c *Context) (string, error) {
	value, err := LookupParameter(n.Name, c)
	if err != nil {
		return "", fmt.Errorf("could not expand %s", n.Name)
	}
	return value, nil
}

func (n ExpansionNode) Vars() []string {
	return []string{ n.Name }
}

type ZeroFillExpansionNode struct {
	Name string
	FieldWidth int
}

func (n ZeroFillExpansionNode) Expand(c *Context) (string, error) {
	value, err := LookupParameter(n.Name, c)
	if err != nil {
		return "", fmt.Errorf("could not expand %s", n.Name)
	}
	numValue, err := strconv.Atoi(value)
	if err != nil {
		return "", fmt.Errorf("could not read %s as integer", value)
	}
	format := fmt.Sprintf("%%0%dd", n.FieldWidth)
	return fmt.Sprintf(format, numValue), nil
}


func (n ZeroFillExpansionNode) Vars() []string {
	return []string{ n.Name }
}


