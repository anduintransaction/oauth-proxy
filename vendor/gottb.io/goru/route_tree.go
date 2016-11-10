package goru

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type routeItemType int

func (t routeItemType) String() string {
	switch t {
	case routeItemRoot:
		return "root"
	case routeItemStatic:
		return "static"
	case routeItemSingle:
		return "single"
	case routeItemRegex:
		return "regex"
	case routeItemMulti:
		return "multi"
	default:
		return "unknown"
	}
}

const (
	routeItemRoot routeItemType = iota
	routeItemStatic
	routeItemRegex
	routeItemSingle
	routeItemMulti
)

type routeTree struct {
	root *routeItem
}

func newRouteTree() *routeTree {
	return &routeTree{
		root: &routeItem{
			itemType: routeItemRoot,
		},
	}
}

func (t *routeTree) add(verb Verb, route string, handler Handler) {
	newRouteItems := t.breakRouteWhenAdd(route)
	leaf := t.walkAdd(newRouteItems, t.root)
	if leaf.handlers == nil {
		leaf.handlers = make(map[Verb]Handler)
	}
	leaf.handlers[verb] = handler
}

func (t *routeTree) get(verb Verb, route string) (h Handler, args map[string]string) {
	args = make(map[string]string)
	segments := t.breakRouteWhenGet(route)
	leaf := t.walkGet(segments, t.root, args)
	if leaf == nil {
		return
	}
	var ok bool
	if h, ok = leaf.handlers[verb]; ok {
		return
	} else if h, ok = leaf.handlers[ANY]; ok {
		return
	}
	return
}

func (t *routeTree) String() string {
	return t.walkPrint("", t.root, 0)
}

func (t *routeTree) walkAdd(newRouteItems []*routeItem, current *routeItem) *routeItem {
	if len(newRouteItems) == 0 {
		return current
	}
	first := newRouteItems[0]
	matched := t.findWhenAdd(first, current.children)
	if matched != nil {
		return t.walkAdd(newRouteItems[1:], matched)
	}
	current.children = append(current.children, first)
	return t.walkAdd(newRouteItems[1:], first)
}

func (t *routeTree) findWhenAdd(item *routeItem, children []*routeItem) *routeItem {
	for _, child := range children {
		if item.matchAdd(child) {
			return child
		}
	}
	return nil
}

func (t *routeTree) walkPrint(buffer string, current *routeItem, level int) string {
	buffer += fmt.Sprintf("%s- %s\n", strings.Repeat("\t", level), current)
	for _, child := range current.children {
		buffer = t.walkPrint(buffer, child, level+1)
	}
	return buffer
}

func (t *routeTree) walkGet(segments []string, current *routeItem, args map[string]string) *routeItem {
	if len(segments) == 0 {
		if len(current.children) == 0 || len(current.handlers) > 0 {
			return current
		}
		for _, child := range current.children {
			if child.itemType == routeItemMulti {
				return child
			}
		}
		return nil
	}
	first := segments[0]
	matches := t.findWhenGet(first, current.children)
	for _, match := range matches {
		remain := segments[1:]
		t.addArg(first, remain, match, args)
		if match.itemType == routeItemMulti {
			remain = nil
		}
		leaf := t.walkGet(remain, match, args)
		if leaf != nil {
			return leaf
		}
		t.removeArg(match, args)
	}
	return nil
}

func (t *routeTree) findWhenGet(segment string, children []*routeItem) []*routeItem {
	matches := []*routeItem{}
	for _, child := range children {
		if child.matchGet(segment) {
			matches = append(matches, child)
		}
	}
	sort.Stable(routeItemSlice(matches))
	return matches
}

func (t *routeTree) addArg(segment string, remain []string, match *routeItem, args map[string]string) {
	switch match.itemType {
	case routeItemSingle, routeItemRegex:
		args[match.value] = segment
	case routeItemMulti:
		args[match.value] = strings.Join(append([]string{segment}, remain...), "/")
	}
}

func (t *routeTree) removeArg(match *routeItem, args map[string]string) {
	delete(args, match.value)
}

var slashRegex = regexp.MustCompile("/+")
var itemRegex = regexp.MustCompile("^\\$([^<]+)<(.+)>$")

func (t *routeTree) breakRouteWhenAdd(route string) []*routeItem {
	trimmed := slashRegex.ReplaceAllString(strings.Trim(route, "/"), "/")
	if trimmed == "" {
		return []*routeItem{}
	}
	pieces := strings.Split(trimmed, "/")
	routeItems := []*routeItem{}
	for i, piece := range pieces {
		newItem := &routeItem{}
		switch piece[0] {
		case ':':
			newItem.itemType = routeItemSingle
			newItem.value = piece[1:]
		case '$':
			newItem.itemType = routeItemRegex
			matches := itemRegex.FindStringSubmatch(piece)
			if len(matches) != 3 {
				panic("part is not a valid regex: " + piece)
			}
			newItem.value = matches[1]
			regexContent := matches[2]
			if !(strings.HasPrefix(regexContent, "^")) {
				regexContent = "^" + regexContent
			}
			if !(strings.HasSuffix(regexContent, "$")) {
				regexContent = regexContent + "$"
			}
			r, err := regexp.Compile(regexContent)
			if err != nil {
				panic("part is not a valid regex: " + piece)
			}
			newItem.regex = r
		case '*':
			if i != len(pieces)-1 {
				panic("part spans multi segments must be at the end of the route")
			}
			newItem.itemType = routeItemMulti
			newItem.value = piece[1:]
		default:
			newItem.itemType = routeItemStatic
			newItem.value = piece
		}
		routeItems = append(routeItems, newItem)
	}
	return routeItems
}

func (t *routeTree) breakRouteWhenGet(route string) []string {
	trimmed := slashRegex.ReplaceAllString(strings.Trim(route, "/"), "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

type routeItem struct {
	handlers map[Verb]Handler
	children []*routeItem
	itemType routeItemType
	value    string
	regex    *regexp.Regexp
}

func (item *routeItem) matchAdd(another *routeItem) bool {
	if item.itemType != another.itemType || item.value != another.value {
		return false
	}
	if item.itemType == routeItemRegex {
		return item.regex.String() == another.regex.String()
	}
	return true
}

func (item *routeItem) matchGet(segment string) bool {
	switch item.itemType {
	case routeItemStatic:
		return item.value == segment
	case routeItemSingle, routeItemMulti:
		return true
	case routeItemRegex:
		return item.regex.MatchString(segment)
	default:
		return false
	}
}

func (item *routeItem) String() string {
	switch item.itemType {
	case routeItemRoot:
		return fmt.Sprintf("[root] => %v", item.handlers)
	case routeItemStatic:
		return fmt.Sprintf("%s => %v", item.value, item.handlers)
	case routeItemSingle:
		return fmt.Sprintf(":%s => %v", item.value, item.handlers)
	case routeItemRegex:
		return fmt.Sprintf("$%s<%s> => %v", item.value, item.regex, item.handlers)
	case routeItemMulti:
		return fmt.Sprintf("*%s => %v", item.value, item.handlers)
	default:
		return "unknown"
	}

}

type routeItemSlice []*routeItem

func (s routeItemSlice) Len() int {
	return len(s)
}

func (s routeItemSlice) Less(i, j int) bool {
	return s[i].itemType < s[j].itemType
}

func (s routeItemSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
