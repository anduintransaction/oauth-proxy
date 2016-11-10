package toml

import "io"

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

func (decoder *Decoder) Decode() (*Value, error) {
	l := &lexer{}
	tokens, err := l.lex(decoder.r)
	if err != nil {
		return nil, err
	}
	p := &parser{}
	return p.parse(tokens)
}
