package handlers

import (
	"net/http"
	"strings"

	"gottb.io/goru"
	"gottb.io/goru/packer"
)

type Assets struct {
	Name     string
	Packs    map[string]*packer.Pack
	NotFound goru.Handler
}

func (h *Assets) Handle(c *goru.Context) {
	if h.NotFound == nil {
		h.NotFound = goru.NetHTTPHandlerFunc(http.NotFound)
	}
	path, ok := c.Params[h.Name]
	if !ok {
		h.NotFound.Handle(c)
		return
	}
	pieces := strings.Split(path, "/")
	if len(pieces) < 2 {
		h.NotFound.Handle(c)
		return
	}
	packName := pieces[0]
	file := strings.Join(pieces[1:], "/")
	pack, ok := h.Packs[packName]
	if !ok {
		h.NotFound.Handle(c)
		return
	}
	f, err := pack.Open(file)
	if err != nil {
		h.NotFound.Handle(c)
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		h.NotFound.Handle(c)
		return
	}
	http.ServeContent(c.ResponseWriter, c.Request, stat.Name(), stat.ModTime(), f)
}
