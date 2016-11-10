package toml

import (
	"fmt"
	"regexp"

	"gottb.io/goru/errors"
)

/*
Grammar:

toml := <table> <toml>?
table := <table_name>? <table_definitions>
table_name := <table_name_single> | <table_name_array>
table_name_single := "[" <keys> "]"
table_name_array := "[[" <keys> "]]"
keys := <key> ("." <keys>)?
key := <bare> | <bare-int> | <quoted>
table_definitions := <table_definition> <table_definitions>?
table_definition := <lhs> "=" <rhs>
lhs := <key>
rhs := <int> | <float> | <boolean> | <datetime> | <string> | <array> | <inline_table>
array := "[" <primitive> ("," <primitive>)* "]"
primitive := <int> | <float> | <boolean> | <datetime> | <string> | <array>
inline_table := "{" <inline_definitions> "}"
inline_definitions := <table_definition> (, <inline_definitions>)?
*/

type ParserError struct {
	reason string
	tok    *token
}

func parserError(reason string, tok *token) *ParserError {
	return &ParserError{
		reason: reason,
		tok:    tok,
	}
}

func (err *ParserError) Error() string {
	return fmt.Sprintf("parser error: %s, line: %d, column: %d, token: %s", err.reason, err.tok.line, err.tok.column, err.tok.Value())
}

type parser struct {
	tokens       []*token
	currentToken int
	root         *Value
	currentTable []string
	arrayTable   bool
}

func (p *parser) parse(tokens []*token) (*Value, error) {
	p.tokens = tokens
	p.currentToken = 0
	p.root = newTable()
	err := p.parseToml()
	if err != nil {
		return nil, err
	}
	return p.root, nil
}

func (p *parser) parseToml() error {
	if p.end() {
		return nil
	}
	err := p.parseTable()
	if err != nil {
		return err
	}
	return p.parseToml()
}

func (p *parser) parseTable() error {
	p.currentTable = []string{}
	p.arrayTable = false
	current := p.current()
	switch current.tokenType {
	case tokenSpecialChar:
		if current.runeValue != '[' {
			return errors.Wrap(parserError("invalid table open charactor", current))
		}
		next := p.next()
		if next == nil {
			return errors.Wrap(parserError("unexpected table name ending", current))
		}
		p.consume()
		current = next
		var err error
		if current.runeValue == '[' {
			err = p.parseTableNameArray()
			p.arrayTable = true
		} else {
			err = p.parseTableNameSingle()
			p.arrayTable = false
		}
		if err != nil {
			return err
		}
		p.consume()
	}
	currentToml, err := p.root.make(p.currentTable, p.arrayTable)
	if err != nil {
		return errors.Wrap(parserError(err.Error(), current))
	}
	values := make(map[string]*Value)
	err = p.parseTableDefinitions(values)
	if err != nil {
		return err
	}
	for k, v := range values {
		currentToml.set(k, v)
	}
	return nil
}

func (p *parser) parseTableNameSingle() error {
	return p.parseKeys()
}

func (p *parser) parseTableNameArray() error {
	p.consume()
	err := p.parseKeys()
	if err != nil {
		return err
	}
	last := p.consume()
	current := p.current()
	if current == nil {
		return errors.Wrap(parserError("unexpected table name ending", last))
	}
	if current.runeValue != ']' {
		return errors.Wrap(parserError("invalid table close charactor", current))
	}
	return nil
}

func (p *parser) parseKeys() error {
	key, err := p.parseKey()
	if err != nil {
		return err
	}
	p.currentTable = append(p.currentTable, key)
	last := p.prev()
	current := p.current()
	if current == nil {
		return errors.Wrap(parserError("unexpected table name ending", last))
	}
	if current.tokenType != tokenSpecialChar {
		return errors.Wrap(parserError("invalid table name", current))
	}
	if current.runeValue != ']' && current.runeValue != '.' {
		return errors.Wrap(parserError("invalid table name", current))
	}
	if current.runeValue == ']' {
		return nil
	}
	p.consume()
	return p.parseKeys()
}

func (p *parser) parseKey() (string, error) {
	current := p.consume()
	if !p.isValidKey(current) {
		return "", errors.Wrap(parserError("invalid key", current))
	}
	return current.stringValue, nil
}

func (p *parser) parseTableDefinitions(values map[string]*Value) error {
	if p.tableEnd() {
		return nil
	}
	lhs, rhs, err := p.parseTableDefinition()
	if err != nil {
		return err
	}
	values[lhs] = rhs
	return p.parseTableDefinitions(values)
}

func (p *parser) parseTableDefinition() (string, *Value, error) {
	lhs, err := p.parseLhs()
	if err != nil {
		return "", nil, err
	}
	current := p.consume()
	if current == nil {
		return "", nil, errors.Wrap(parserError("unexpected EOF", p.prev()))
	}
	if current.runeValue != '=' {
		return "", nil, errors.Wrap(parserError("invalid character in definition", current))
	}
	rhs, err := p.parseRhs()
	if err != nil {
		return "", nil, err
	}
	return lhs, rhs, nil
}

func (p *parser) parseLhs() (string, error) {
	return p.parseKey()
}

func (p *parser) parseRhs() (*Value, error) {
	current := p.consume()
	if current == nil {
		return nil, errors.Wrap(parserError("unexpected EOF", p.prev()))
	}
	switch current.tokenType {
	case tokenString:
		return newString(current.stringValue), nil
	case tokenInt:
		return newInt(current.intValue), nil
	case tokenFloat:
		return newFloat(current.floatValue), nil
	case tokenBoolean:
		return newBool(current.boolValue), nil
	case tokenDatetime:
		return newTime(current.timeValue), nil
	case tokenSpecialChar:
		switch current.runeValue {
		case '{':
			val, err := p.parseInlineTable()
			if err != nil {
				return nil, err
			}
			current = p.consume()
			if current == nil {
				return nil, errors.Wrap(parserError("unexpected EOF", p.prev()))
			}
			if current.runeValue != '}' {
				return nil, errors.Wrap(parserError("invalid inline table close charactor", current))
			}
			return val, nil
		case '[':
			val, err := p.parseInlineArray()
			if err != nil {
				return nil, err
			}
			current = p.consume()
			if current == nil {
				return nil, errors.Wrap(parserError("unexpected EOF", p.prev()))
			}
			if current.runeValue != ']' {
				return nil, errors.Wrap(parserError("invalid inline array close charactor", current))
			}
			return val, nil
		default:
			return nil, errors.Wrap(parserError("unexpected token: "+string(current.runeValue), current))
		}
	default:
		return nil, errors.Wrap(parserError("unexpected type", current))
	}
}

func (p *parser) parseInlineTable() (*Value, error) {
	if p.end() {
		return nil, errors.Wrap(parserError("unexpected EOF", p.prev()))
	}
	toml := newTable()
	values := make(map[string]*Value)
	err := p.parseInlineTableDefinitions(values)
	if err != nil {
		return nil, err
	}
	for k, v := range values {
		toml.set(k, v)
	}
	return toml, nil
}

func (p *parser) parseInlineArray() (*Value, error) {
	if p.end() {
		return nil, errors.Wrap(parserError("unexpected EOF", p.prev()))
	}
	toml := newArray()
	values, err := p.parseInlineArrayPrimitives()
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		value.inArray = true
		toml.add(value)
	}
	return toml, nil
}

func (p *parser) parseInlineTableDefinitions(values map[string]*Value) error {
	if p.current().runeValue == '}' {
		return nil
	}
	lhs, rhs, err := p.parseTableDefinition()
	if err != nil {
		return err
	}
	values[lhs] = rhs
	if p.end() {
		return errors.Wrap(parserError("unexpected EOF", p.prev()))
	}
	current := p.current()
	if current.runeValue == ',' {
		p.consume()
		return p.parseInlineTableDefinitions(values)
	}
	if current.runeValue != '}' {
		return errors.Wrap(parserError("unexpected token", current))
	}
	return nil
}

func (p *parser) parseInlineArrayPrimitives() ([]*Value, error) {
	if p.current().runeValue == ']' {
		return []*Value{}, nil
	}
	value, err := p.parseRhs()
	if err != nil {
		return nil, err
	}
	if p.end() {
		return nil, errors.Wrap(parserError("unexpected EOF", p.prev()))
	}
	current := p.current()
	if current.runeValue == ',' {
		p.consume()
		next, err := p.parseInlineArrayPrimitives()
		if err != nil {
			return nil, err
		}
		return append([]*Value{value}, next...), nil
	}
	if current.runeValue != ']' {
		return nil, errors.Wrap(parserError("unexpected token", current))
	}
	return []*Value{value}, nil
}

func (p *parser) tableEnd() bool {
	if p.end() {
		return true
	}
	return p.current().runeValue == '['
}

func (p *parser) prev() *token {
	if p.currentToken == 0 || len(p.tokens) == 0 {
		return nil
	}
	return p.tokens[p.currentToken-1]
}

func (p *parser) current() *token {
	if p.currentToken >= len(p.tokens) {
		return nil
	} else {
		return p.tokens[p.currentToken]
	}
}

func (p *parser) next() *token {
	if p.currentToken+1 >= len(p.tokens) {
		return nil
	}
	return p.tokens[p.currentToken+1]
}

func (p *parser) consume() *token {
	tok := p.current()
	if tok != nil {
		p.currentToken++
	}
	return tok
}

func (p *parser) end() bool {
	return p.currentToken >= len(p.tokens)
}

func (p *parser) isValidKey(tok *token) bool {
	if tok.tokenType == tokenBare || tok.tokenType == tokenString {
		return true
	}
	if tok.tokenType == tokenInt {
		return p.isValidBare(tok.stringValue)
	}
	return false
}

var validBare = regexp.MustCompile("^[a-zA-Z0-9_\\-]+$")

func (p *parser) isValidBare(str string) bool {
	return validBare.MatchString(str)
}
