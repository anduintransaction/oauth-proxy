package toml

import (
	"os"
	"testing"
)

func TestMerge(t *testing.T) {
	r, err := os.Open("example-v0.4.0.toml")
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(r)
	root, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}

	extraR, err := os.Open("extra.toml")
	if err != nil {
		t.Fatal(err)
	}
	extraDecoder := NewDecoder(extraR)
	extraRoot, err := extraDecoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	err = Merge(root, extraRoot)
	if err != nil {
		t.Fatal(err)
	}
	if v, err := root.Get("table.key"); v.(*Value).value.(string) != "value1" {
		t.Fatal("table.key", err, v)
	}
	if v, err := root.Get("table.key2"); v.(*Value).value.(string) != "value2" {
		t.Fatal("table.key2", err, v)
	}
	if v, err := root.Get("table.subtable.key"); v.(*Value).value.(string) != "another value1" {
		t.Fatal("table.subtable.key", err, v)
	}
	if v, err := root.Get("array.key1"); err != nil {
		t.Fatal("array.key1", err)
	} else {
		var arr []int
		Unmarshal(v, &arr)
		if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 || arr[4] != 5 {
			t.Fatal("array.key1", v, arr)
		}
	}
	if v, err := root.Get("fruit"); err != nil {
		t.Fatal("fruit", err)
	} else {
		var arr []struct {
			Name string `config:"name"`
		}
		Unmarshal(v, &arr)
		if len(arr) != 1 || arr[0].Name != "Mango" {
			t.Fatal("fruit", v, arr)
		}
	}
}
