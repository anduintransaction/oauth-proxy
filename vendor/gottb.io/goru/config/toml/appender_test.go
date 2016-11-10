package toml

import (
	"os"
	"testing"
	"time"
)

func TestAppend(t *testing.T) {
	r, err := os.Open("example-v0.4.0.toml")
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(r)
	root, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	Append(root, "array.key1", 4, 5, 6)
	if v, err := root.Get("array.key1"); err != nil {
		t.Fatal("array.key1")
	} else {
		var arr []int
		Unmarshal(v, &arr)
		if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 || arr[4] != 5 || arr[5] != 6 {
			t.Fatal("array.key1", err, v)
		}
	}
	Append(root, "array.key2", "blue", "pink")
	if v, err := root.Get("array.key2"); err != nil {
		t.Fatal("array.key2")
	} else {
		var arr []string
		Unmarshal(v, &arr)
		if arr[0] != "red" || arr[1] != "yellow" || arr[2] != "green" || arr[3] != "blue" || arr[4] != "pink" {
			t.Fatal("array.key2", err, v)
		}
	}
	Append(root, "array.new.key", 1.1, 2.2, 3.3)
	if v, err := root.Get("array.new.key"); err != nil {
		t.Fatal("array.new.key")
	} else {
		var arr []float64
		Unmarshal(v, &arr)
		if arr[0]-1.1 > 1e-5 || arr[1]-2.2 > 1e-5 || arr[2]-3.3 > 1e-5 {
			t.Fatal("array.new.key", err, v)
		}
	}
	now := time.Now()
	Append(root, "array.time", now, now)
	if v, err := root.Get("array.time"); err != nil {
		t.Fatal("array.time")
	} else {
		var arr []time.Time
		Unmarshal(v, &arr)
	}
	Append(root, "array.tensor", [][]int{[]int{1, 2, 3}, []int{4, 5}}, [][]int{[]int{6, 7}})
	if v, err := root.Get("array.tensor"); err != nil {
		t.Fatal("array.tensor")
	} else {
		var arr [][][]int
		Unmarshal(v, &arr)
		if arr[0][0][0] != 1 || arr[0][0][1] != 2 || arr[0][0][2] != 3 ||
			arr[0][1][0] != 4 || arr[0][1][1] != 5 || arr[1][0][0] != 6 ||
			arr[1][0][1] != 7 {
			t.Fatal("array.tensor", err, v)
		}
	}
	Append(root, "products", &Product{"product1", 1234, "yellow"})
	if v, err := root.Get("products"); err != nil {
		t.Fatal("products", err)
	} else {
		var arr []*Product
		err = Unmarshal(v, &arr)
		if err != nil || arr[0].Name != "Hammer" || arr[0].Sku != 738594937 ||
			arr[2].Name != "Nail" || arr[2].Sku != 284758393 || arr[2].Color != "gray" ||
			arr[3].Name != "product1" || arr[3].Sku != 1234 || arr[3].Color != "yellow" {
			t.Fatal("products", err, arr)
		}
	}

}

type Product struct {
	Name  string `config:"name"`
	Sku   int    `config:"sku"`
	Color string `config:"color"`
}
