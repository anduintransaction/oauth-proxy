package toml

import (
	"os"
	"testing"
	"time"

	"gottb.io/goru/errors"
)

func TestUpdate(t *testing.T) {
	r, err := os.Open("example-v0.4.0.toml")
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(r)
	root, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	Update(root, "table.key", 9999)
	if v, err := root.Get("table.key"); err != nil || v.(*Value).value.(int64) != 9999 {
		t.Fatal("table.key", err, v)
	}
	Update(root, "table.anotherKey", 3.1415)
	if v, err := root.Get("table.anotherKey"); err != nil || v.(*Value).value.(float64) != 3.1415 {
		t.Fatal("table.anotherKey", err, v)
	}
	Update(root, "table.subtable2.key", "value")
	if v, err := root.Get("table.subtable2.key"); err != nil || v.(*Value).value.(string) != "value" {
		t.Fatal("table.subtable2.key", err, v)
	}
	Update(root, "table.subtable2", true)
	if v, err := root.Get("table.subtable2"); err != nil || !v.(*Value).value.(bool) {
		t.Fatal("table.subtable2", err, v)
	}
	now := time.Now()
	Update(root, "table.subtable2", now)
	if v, err := root.Get("table.subtable2"); err != nil || v.(*Value).value.(time.Time) != now {
		t.Fatal("table.subtable2", err, v)
	}
	Update(root, "array.key1.0", 9999)
	if v, err := root.Get("array.key1.0"); err != nil || v.(*Value).value.(int64) != 9999 {
		t.Fatal("array.key1.0", err, v)
	}
	if err = Update(root, "array.key1.9", "value"); !errors.Is(err, ErrNotFound) {
		t.Fatal("array.key1.9", err)
	}
	if err = Update(root, "array.key1.0", "value"); !errors.Is(err, ErrType) {
		t.Fatal("array.key1.0", err)
	}
	Update(root, "array.newKey", []int{1, 2, 3, 4})
	if v, err := root.Get("array.newKey"); err != nil {
		t.Fatal("array.newKey", err)
	} else {
		var arr []int
		Unmarshal(v, &arr)
		if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 {
			t.Fatal("array.newKey", err, v)
		}
	}
	Update(root, "array.datetime", []time.Time{now, now})
	if v, err := root.Get("array.datetime"); err != nil {
		t.Fatal("array.datetime", err)
	} else {
		var arr []time.Time
		Unmarshal(v, &arr)
		if arr[0] != now || arr[1] != now {
			t.Fatal("array.datetime", err, v)
		}
	}
	Update(root, "array.matrix", [][]string{
		[]string{"1", "2", "3"},
		[]string{"4", "5"},
	})
	if v, err := root.Get("array.matrix"); err != nil {
		t.Fatal("array.matrix", err)
	} else {
		var matrix [][]string
		Unmarshal(v, &matrix)
		if matrix[0][0] != "1" || matrix[0][1] != "2" || matrix[0][2] != "3" || matrix[1][0] != "4" || matrix[1][1] != "5" {
			t.Fatal("array.matrix", err, v)
		}
	}
	Update(root, "table.map", map[string]int{
		"one": 1,
		"two": 2,
	})
	if v, err := root.Get("table.map"); err != nil {
		t.Fatal("table.map", err)
	} else {
		var m map[string]int
		Unmarshal(v, &m)
		if m["one"] != 1 || m["two"] != 2 {
			t.Fatal("table.map")
		}
	}
	Update(root, "struct", &struct {
		Name       string `config:"name"`
		Age        int    `config:"age"`
		IgnoreThis string
	}{
		Name:       "Finn",
		Age:        1337,
		IgnoreThis: "traitor",
	})
	if v, err := root.Get("struct"); err != nil {
		t.Fatal("struct", err)
	} else {
		var s struct {
			Name       string `config:"name"`
			Age        int    `config:"age"`
			IgnoreThis string
		}
		Unmarshal(v, &s)
		if s.Name != "Finn" || s.Age != 1337 || s.IgnoreThis != "" {
			t.Fatal("struct")
		}
	}
	Update(root, "author", []*author{
		&author{
			Name: "J.R.R Tolkien",
			Books: []*book{
				&book{"The Lord of the Rings"},
				&book{"The Simarillion"},
			},
		},
		&author{
			Name: "Richard Dawkin",
			Books: []*book{
				&book{"The Selfish Gene"},
			},
		},
	})
	if v, err := root.Get("author"); err != nil {
		t.Fatal("author", err)
	} else {
		var authors []*author
		Unmarshal(v, &authors)
		if authors[0].Name != "J.R.R Tolkien" || authors[0].Books[0].Title != "The Lord of the Rings" ||
			authors[0].Books[1].Title != "The Simarillion" || authors[1].Name != "Richard Dawkin" ||
			authors[1].Books[0].Title != "The Selfish Gene" {
			t.Fatal("author", err, v)
		}
	}
	Update(root, "library", map[string][]*book{
		"fantasy": []*book{
			&book{"The Hobbit"},
			&book{"Children of Hurin"},
		},
		"computer science": []*book{
			&book{"Pattern Recognition and Machine Learning"},
		},
	})
	if v, err := root.Get("library"); err != nil {
		t.Fatal("library", err)
	} else {
		var library map[string][]*book
		Unmarshal(v, &library)
		if library["fantasy"][0].Title != "The Hobbit" || library["fantasy"][1].Title != "Children of Hurin" ||
			library["computer science"][0].Title != "Pattern Recognition and Machine Learning" {
			t.Fatal("library", err, v)
		}
	}
}

type author struct {
	Name  string  `config:"name"`
	Books []*book `config:"books"`
}

type book struct {
	Title string `config:"title"`
}
