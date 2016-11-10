package toml

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gottb.io/goru/errors"
)

const (
	tokenBare = iota
	tokenString
	tokenSpecialChar
	tokenInt
	tokenFloat
	tokenBoolean
	tokenDatetime
)

const (
	stateBegin = iota
	stateComment
	stateQuotedBegin
	stateQuotedInside
	stateQuotedNormalEscape
	stateQuotedUnicodeEscape
	stateQuotedWhitespaceEscape
	stateQuotedEnd
	stateLiteralBegin
	stateLiteralInside
	stateLiteralEnd
	stateComplex
	stateNumber
)

const (
	lineSingle = iota
	lineMulti
)

type LexerError struct {
	reason string
	line   int
	column int
}

func (err *LexerError) Error() string {
	return fmt.Sprintf("lexer error: %s, line: %d, column %d", err.reason, err.line, err.column)
}

type token struct {
	tokenType   int
	runeValue   rune
	intValue    int64
	floatValue  float64
	stringValue string
	boolValue   bool
	timeValue   time.Time
	line        int
	column      int
}

func (tok *token) String() string {
	switch tok.tokenType {
	case tokenBare:
		return "BARE(" + tok.stringValue + ")@(" + fmt.Sprint(tok.line) + ":" + fmt.Sprint(tok.column) + ")"
	case tokenString:
		return "STRING(" + tok.stringValue + ")@(" + fmt.Sprint(tok.line) + ":" + fmt.Sprint(tok.column) + ")"
	case tokenSpecialChar:
		return "SCHAR(" + string(tok.runeValue) + ")@(" + fmt.Sprint(tok.line) + ":" + fmt.Sprint(tok.column) + ")"
	case tokenInt:
		return "INT(" + fmt.Sprint(tok.intValue) + ")@(" + fmt.Sprint(tok.line) + ":" + fmt.Sprint(tok.column) + ")"
	case tokenFloat:
		return "FLOAT(" + fmt.Sprint(tok.floatValue) + ")@(" + fmt.Sprint(tok.line) + ":" + fmt.Sprint(tok.column) + ")"
	case tokenBoolean:
		return "BOOL(" + fmt.Sprint(tok.boolValue) + ")@(" + fmt.Sprint(tok.line) + ":" + fmt.Sprint(tok.column) + ")"
	case tokenDatetime:
		return "DATETIME(" + fmt.Sprint(tok.timeValue) + ")@(" + fmt.Sprint(tok.line) + ":" + fmt.Sprint(tok.column) + ")"
	default:
		return "UNKNOWN"
	}
}

func (tok *token) Value() string {
	switch tok.tokenType {
	case tokenBare, tokenString:
		return tok.stringValue
	case tokenSpecialChar:
		return string(tok.runeValue)
	case tokenInt:
		return fmt.Sprint(tok.intValue)
	case tokenFloat:
		return fmt.Sprint(tok.floatValue)
	case tokenBoolean:
		return fmt.Sprint(tok.boolValue)
	case tokenDatetime:
		return fmt.Sprint(tok.timeValue)
	default:
		return "UNKNOWN"
	}
}

type lexer struct {
	r               *bufio.Reader
	b               *bytes.Buffer
	unicodeBuf      *bytes.Buffer
	unicodeCount    int
	unicodeExpected int
	next            rune
	tokens          []*token
	start           bool
	state           int
	lineState       int
	quoteCount      int
	line            int
	column          int
	tokenLine       int
	tokenColumn     int
}

func (l *lexer) lex(r io.Reader) ([]*token, error) {
	l.r = bufio.NewReader(r)
	l.tokens = []*token{}
	l.start = true
	l.state = stateBegin
	l.line = 1
	var err error
	for {
		err = l.scan()
		if err != nil {
			break
		}
	}
	if !errors.Is(err, io.EOF) {
		return nil, err
	}
	return l.tokens, nil
}

func (l *lexer) scan() error {
	r, _, err := l.r.ReadRune()
	current := l.next
	l.next = r
	if l.start {
		if err != nil {
			return err
		}
		l.start = false
		return nil
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	var lexError error
	if current == '\n' {
		l.line++
		l.column = 0
	}
	l.column++
	switch l.state {
	case stateBegin:
		lexError = l.doBegin(current, l.next)
	case stateComment:
		lexError = l.doComment(current, l.next)
	case stateComplex:
		lexError = l.doComplex(current, l.next)
	case stateQuotedBegin:
		lexError = l.doQuotedBegin(current, l.next)
	case stateQuotedInside:
		lexError = l.doQuotedInside(current, l.next)
	case stateQuotedNormalEscape:
		lexError = l.doQuotedNormalEscape(current, l.next)
	case stateQuotedUnicodeEscape:
		lexError = l.doQuotedUnicodeEscape(current, l.next)
	case stateQuotedWhitespaceEscape:
		lexError = l.doQuotedWhitespaceEscape(current, l.next)
	case stateQuotedEnd:
		lexError = l.doQuotedEnd(current, l.next)
	case stateLiteralBegin:
		lexError = l.doLiteralBegin(current, l.next)
	case stateLiteralInside:
		lexError = l.doLiteralInside(current, l.next)
	case stateLiteralEnd:
		lexError = l.doLiteralEnd(current, l.next)
	case stateNumber:
		lexError = l.doNumber(current, l.next)
	}
	if lexError != nil {
		return lexError
	}
	return err
}

func (l *lexer) doBegin(current, next rune) error {
	switch current {
	case '#':
		l.state = stateComment
	case ' ', '\t', '\n', '\r':
		return nil
	case '[', ']', '{', '}', '.', '=', ',':
		l.tokens = append(l.tokens, &token{
			tokenType: tokenSpecialChar,
			runeValue: current,
			line:      l.line,
			column:    l.column,
		})
	case '"':
		l.state = stateQuotedBegin
		l.quoteCount = 1
		l.b = &bytes.Buffer{}
		l.tokenLine = l.line
		l.tokenColumn = l.column
	case '\'':
		l.state = stateLiteralBegin
		l.quoteCount = 1
		l.b = &bytes.Buffer{}
		l.tokenLine = l.line
		l.tokenColumn = l.column
	case '+':
		l.state = stateNumber
		l.b = &bytes.Buffer{}
		l.tokenLine = l.line
		l.tokenColumn = l.column
		l.doNumber(current, next)
	default:
		if l.isComplexBegin(current) {
			l.b = &bytes.Buffer{}
			l.tokenLine = l.line
			l.tokenColumn = l.column
			l.b.WriteRune(current)
			if l.isComplexInside(l.next) {
				l.state = stateComplex
			} else {
				l.state = stateBegin
				return l.compileComplexToken()
			}
		} else {
			return errors.Wrap(&LexerError{
				reason: "Invalid charactor",
				line:   l.line,
				column: l.column,
			})
		}
	}
	return nil
}

func (l *lexer) doComment(current, next rune) error {
	switch current {
	case '\n':
		l.state = stateBegin
	}
	return nil
}

func (l *lexer) doComplex(current, next rune) error {
	l.b.WriteRune(current)
	if !l.isComplexInside(next) {
		l.state = stateBegin
		return l.compileComplexToken()
	}
	return nil
}

func (l *lexer) doQuotedBegin(current, next rune) error {
	switch {
	case l.quoteCount == 1 && current == '"' && next != '"':
		l.state = stateBegin
		return l.compileStringToken()
	case l.quoteCount == 1 && current == '"' && next == '"':
		l.quoteCount = 2
	case l.quoteCount == 1 && current != '"':
		l.state = stateQuotedInside
		l.lineState = lineSingle
		return l.doQuotedInside(current, next)
	case l.quoteCount == 2:
		l.quoteCount = 3
		if next != '\r' && next != '\n' {
			l.state = stateQuotedInside
			l.lineState = lineMulti
		}
	case l.quoteCount == 3:
		if current == '\n' {
			l.state = stateQuotedInside
			l.lineState = lineMulti
		}
	default:
		return errors.Wrap(&LexerError{
			reason: "wrong quote",
			line:   l.line,
			column: l.column,
		})
	}
	return nil
}

func (l *lexer) doQuotedInside(current, next rune) error {
	if current != '\\' && current != '"' {
		l.b.WriteRune(current)
		return nil
	}
	if current == '"' && l.lineState == lineSingle {
		l.state = stateBegin
		return l.compileStringToken()
	}
	if current == '"' && l.lineState == lineMulti {
		l.state = stateQuotedEnd
		l.quoteCount = 1
		return nil
	}
	if current == '\\' {
		switch next {
		case 'b', 't', 'n', 'f', 'r', '"', '\\':
			l.state = stateQuotedNormalEscape
		case 'u', 'U':
			l.state = stateQuotedUnicodeEscape
		case ' ', '\t', '\r', '\n':
			l.state = stateQuotedWhitespaceEscape
		default:
			return &LexerError{
				reason: "invalid escape sequence",
				line:   l.line,
				column: l.column,
			}
		}
	}
	return nil
}

func (l *lexer) doQuotedNormalEscape(current, next rune) error {
	switch current {
	case 'b':
		l.b.WriteRune('\b')
	case 't':
		l.b.WriteRune('\t')
	case 'r':
		l.b.WriteRune('\r')
	case 'n':
		l.b.WriteRune('\n')
	case 'f':
		l.b.WriteRune('\f')
	case '"':
		l.b.WriteRune('"')
	case '\\':
		l.b.WriteRune('\\')
	}
	l.state = stateQuotedInside
	return nil
}

func (l *lexer) doQuotedUnicodeEscape(current, next rune) error {
	switch current {
	case 'u':
		l.unicodeBuf = &bytes.Buffer{}
		l.unicodeCount = 0
		l.unicodeExpected = 4
	case 'U':
		l.unicodeBuf = &bytes.Buffer{}
		l.unicodeCount = 0
		l.unicodeExpected = 8
	default:
		l.unicodeBuf.WriteRune(current)
		l.unicodeCount++
		if l.unicodeCount >= l.unicodeExpected {
			l.state = stateQuotedInside
			val, err := strconv.ParseInt(l.unicodeBuf.String(), 16, 64)
			if err != nil {
				return errors.Wrap(&LexerError{
					reason: "invalid unicode format: " + l.unicodeBuf.String(),
					line:   l.line,
					column: l.column,
				})
			}
			u := rune(val)
			if u > unicode.MaxRune || u == unicode.ReplacementChar {
				return errors.Wrap(&LexerError{
					reason: "invalid unicode code point: " + l.unicodeBuf.String(),
					line:   l.line,
					column: l.column,
				})
			}
			l.b.WriteRune(u)
		}
	}
	return nil
}

func (l *lexer) doQuotedWhitespaceEscape(current, next rune) error {
	if next != ' ' && next != '\t' && next != '\r' && next != '\n' {
		l.state = stateQuotedInside
	}
	return nil
}

func (l *lexer) doQuotedEnd(current, next rune) error {
	switch {
	case l.quoteCount == 1 && current == '"' && next == '"':
		l.quoteCount = 2
		return nil
	case l.quoteCount == 2:
		l.state = stateBegin
		return l.compileStringToken()
	default:
		return errors.Wrap(&LexerError{
			reason: "invalid end quote",
			line:   l.line,
			column: l.column,
		})
	}
}

func (l *lexer) doLiteralBegin(current, next rune) error {
	switch {
	case l.quoteCount == 1 && current == '\'' && next != '\'':
		l.state = stateBegin
		return l.compileStringToken()
	case l.quoteCount == 1 && current == '\'' && next == '\'':
		l.quoteCount = 2
	case l.quoteCount == 1 && current != '\'':
		l.state = stateLiteralInside
		l.lineState = lineSingle
		return l.doLiteralInside(current, next)
	case l.quoteCount == 2:
		l.quoteCount = 3
		if next != '\r' && next != '\n' {
			l.state = stateLiteralInside
			l.lineState = lineMulti
		}
	case l.quoteCount == 3:
		if current == '\n' {
			l.state = stateLiteralInside
			l.lineState = lineMulti
		}
	default:
		return errors.Wrap(&LexerError{
			reason: "wrong quote",
			line:   l.line,
			column: l.column,
		})
	}
	return nil
}

func (l *lexer) doLiteralInside(current, next rune) error {
	if current != '\'' {
		l.b.WriteRune(current)
		return nil
	}
	if current == '\'' && l.lineState == lineSingle {
		l.state = stateBegin
		return l.compileStringToken()
	}
	if current == '\'' && l.lineState == lineMulti {
		if next == '\'' {
			l.state = stateLiteralEnd
			l.quoteCount = 1
		} else {
			l.b.WriteRune(current)
		}
	}
	return nil
}

func (l *lexer) doLiteralEnd(current, next rune) error {
	switch {
	case l.quoteCount == 1 && current == '\'' && next == '\'':
		l.quoteCount = 2
		return nil
	case l.quoteCount == 2:
		l.state = stateBegin
		return l.compileStringToken()
	default:
		return errors.Wrap(&LexerError{
			reason: "invalid end quote literal",
			line:   l.line,
			column: l.column,
		})
	}
}

func (l *lexer) doNumber(current, next rune) error {
	l.b.WriteRune(current)
	if !l.isNumber(next) {
		l.state = stateBegin
		orig := l.b.String()
		maybeNumber := strings.Replace(orig, "_", "", -1)
		maybeInt, err := strconv.ParseInt(maybeNumber, 10, 64)
		if err == nil {
			l.tokens = append(l.tokens, &token{
				tokenType:   tokenInt,
				intValue:    maybeInt,
				stringValue: orig,
				line:        l.tokenLine,
				column:      l.tokenColumn,
			})
			return nil
		}
		maybeFloat, err := strconv.ParseFloat(maybeNumber, 64)
		if err == nil {
			l.tokens = append(l.tokens, &token{
				tokenType:  tokenFloat,
				floatValue: maybeFloat,
				line:       l.tokenLine,
				column:     l.tokenColumn,
			})
			return nil
		}
		return errors.Wrap(&LexerError{
			reason: "invalid number",
			line:   l.line,
			column: l.column,
		})
	}
	return nil
}

func (l *lexer) compileComplexToken() error {
	orig := l.b.String()
	maybeNumber := strings.Replace(orig, "_", "", -1)
	maybeInt, err := strconv.ParseInt(maybeNumber, 10, 64)
	if err == nil {
		l.tokens = append(l.tokens, &token{
			tokenType:   tokenInt,
			intValue:    maybeInt,
			stringValue: orig,
			line:        l.tokenLine,
			column:      l.tokenColumn,
		})
		return nil
	}
	maybeFloat, err := strconv.ParseFloat(maybeNumber, 64)
	if err == nil {
		l.tokens = append(l.tokens, &token{
			tokenType:  tokenFloat,
			floatValue: maybeFloat,
			line:       l.tokenLine,
			column:     l.tokenColumn,
		})
		return nil
	}
	if orig == "true" {
		l.tokens = append(l.tokens, &token{
			tokenType: tokenBoolean,
			boolValue: true,
			line:      l.tokenLine,
			column:    l.tokenColumn,
		})
		return nil
	}
	if orig == "false" {
		l.tokens = append(l.tokens, &token{
			tokenType: tokenBoolean,
			boolValue: false,
			line:      l.tokenLine,
			column:    l.tokenColumn,
		})
		return nil
	}
	t, err := time.Parse(time.RFC3339, orig)
	if err == nil {
		l.tokens = append(l.tokens, &token{
			tokenType: tokenDatetime,
			timeValue: t,
			line:      l.tokenLine,
			column:    l.tokenColumn,
		})
		return nil
	}
	if strings.ContainsRune(orig, ':') {
		return errors.Wrap(&LexerError{
			reason: "Invalid bare string: " + orig,
			line:   l.line,
			column: l.column,
		})
	}
	l.tokens = append(l.tokens, &token{
		tokenType:   tokenBare,
		stringValue: orig,
		line:        l.tokenLine,
		column:      l.tokenColumn,
	})
	return nil
}

func (l *lexer) compileStringToken() error {
	l.tokens = append(l.tokens, &token{
		tokenType:   tokenString,
		stringValue: l.b.String(),
		line:        l.tokenLine,
		column:      l.tokenColumn,
	})
	return nil
}

func (l *lexer) isComplexBegin(r rune) bool {
	return r >= 'a' && r <= 'z' ||
		r >= 'A' && r <= 'Z' ||
		r >= '0' && r <= '9' ||
		r == '_' || r == '-'
}

func (l *lexer) isComplexInside(r rune) bool {
	return r >= 'a' && r <= 'z' ||
		r >= 'A' && r <= 'Z' ||
		r >= '0' && r <= '9' ||
		r == '+' || r == '_' || r == '-' || r == ':' || (l.lastToken() != nil && l.lastToken().runeValue == '=' && r == '.')
}

func (l *lexer) isNumber(r rune) bool {
	return r >= '0' && r <= '9' || r == '.' || r == '_' || r == '+' || r == '-' || r == 'e' || r == 'E'
}

func (l *lexer) lastToken() *token {
	if len(l.tokens) == 0 {
		return nil
	}
	return l.tokens[len(l.tokens)-1]
}
