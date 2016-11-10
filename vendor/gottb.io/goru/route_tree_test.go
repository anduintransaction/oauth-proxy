package goru

import "testing"

type fakeHandler struct {
	value string
}

func (f *fakeHandler) Handle(c *Context) {
}

func TestRouteTree(t *testing.T) {
	tree := newRouteTree()
	tree.add(GET, "/", &fakeHandler{"root-get"})
	tree.add(ANY, "/", &fakeHandler{"root-any"})
	tree.add(GET, "/user", &fakeHandler{"user-list"})
	tree.add(GET, "/user/:id", &fakeHandler{"user-view"})
	tree.add(GET, "/user/:id/edit", &fakeHandler{"user-edit"})
	tree.add(GET, "/user/goru", &fakeHandler{"user-goru"})
	tree.add(GET, "/user/gottb/goru", &fakeHandler{"user-gottb"})

	tree.add(GET, "/post", &fakeHandler{"post-list"})
	tree.add(GET, "/post/$id<\\d+>", &fakeHandler{"post-view"})
	tree.add(GET, "/post/$id<\\d+>/edit", &fakeHandler{"post-edit"})
	tree.add(GET, "/post/$slug<[a-zA-Z0-9\\-]+>", &fakeHandler{"post-slug"})
	tree.add(GET, "/post/:query", &fakeHandler{"post-query"})
	tree.add(GET, "/post/*id-list", &fakeHandler{"post-view-multi"})

	if h, args := tree.get(GET, "/"); h == nil || h.(*fakeHandler).value != "root-get" || len(args) != 0 {
		t.Fatal("root-get: ", h, args)
	}
	if h, args := tree.get(POST, "/"); h == nil || h.(*fakeHandler).value != "root-any" || len(args) != 0 {
		t.Fatal("root-any: ", h, args)
	}
	if h, args := tree.get(GET, "/user"); h == nil || h.(*fakeHandler).value != "user-list" || len(args) != 0 {
		t.Fatal("user-list: ", h, args)
	}
	if h, args := tree.get(POST, "/user"); h != nil {
		t.Fatal("user-list-post: ", h, args)
	}
	if h, args := tree.get(GET, "/user/rinoa"); h == nil || h.(*fakeHandler).value != "user-view" || args["id"] != "rinoa" {
		t.Fatal("user-view: ", h, args)
	}
	if h, args := tree.get(GET, "/user/rinoa/edit"); h == nil || h.(*fakeHandler).value != "user-edit" || args["id"] != "rinoa" {
		t.Fatal("user-edit: ", h, args)
	}
	if h, args := tree.get(GET, "/user/goru"); h == nil || h.(*fakeHandler).value != "user-goru" || len(args) != 0 {
		t.Fatal("user-goru: ", h, args)
	}
	if h, args := tree.get(GET, "/user/gottb"); h == nil || h.(*fakeHandler).value != "user-view" || args["id"] != "gottb" {
		t.Fatal("user-view: ", h, args)
	}
	if h, args := tree.get(GET, "/user/gottb/goru"); h == nil || h.(*fakeHandler).value != "user-gottb" || len(args) != 0 {
		t.Fatal("user-gottb: ", h, args)
	}
	if h, args := tree.get(GET, "/user/gottb/edit"); h == nil || h.(*fakeHandler).value != "user-edit" || args["id"] != "gottb" {
		t.Fatal("user-edit: ", h, args)
	}
	if h, args := tree.get(GET, "/user/gottb/delete"); h != nil {
		t.Fatal("user-delete: ", h, args)
	}
	if h, args := tree.get(GET, "/post"); h == nil || h.(*fakeHandler).value != "post-list" || len(args) != 0 {
		t.Fatal("post-list: ", h, args)
	}
	if h, args := tree.get(GET, "/post/1337"); h == nil || h.(*fakeHandler).value != "post-view" || args["id"] != "1337" {
		t.Fatal("post-view: ", h, args)
	}
	if h, args := tree.get(GET, "/post/1337/edit"); h == nil || h.(*fakeHandler).value != "post-edit" || args["id"] != "1337" {
		t.Fatal("post-edit: ", h, args)
	}
	if h, args := tree.get(GET, "/post/goru-is-fun"); h == nil || h.(*fakeHandler).value != "post-slug" || args["slug"] != "goru-is-fun" {
		t.Fatal("post-slug: ", h, args)
	}
	if h, args := tree.get(GET, "/post/THIS! IS! SPARTA!!!!"); h == nil || h.(*fakeHandler).value != "post-query" || args["query"] != "THIS! IS! SPARTA!!!!" {
		t.Fatal("post-query: ", h, args)
	}
	if h, args := tree.get(GET, "/post/1/2/3/4"); h == nil || h.(*fakeHandler).value != "post-view-multi" || args["id-list"] != "1/2/3/4" {
		t.Fatal("post-view-multi: ", h, args)
	}
}
