package utils

type StringSet map[string]struct{}

func NewStringSet(values []string) StringSet {
	s := make(StringSet)
	for _, value := range values {
		s.Add(value)
	}
	return s
}

func (s StringSet) Add(value string) {
	s[value] = struct{}{}
}

func (s StringSet) Remove(value string) {
	delete(s, value)
}

func (s StringSet) Has(value string) bool {
	_, ok := s[value]
	return ok
}
