package toml

import (
	"os"
	"testing"
	"time"
)

func TestUnmarshal(t *testing.T) {
	r, err := os.Open("example-v0.4.0.toml")
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(r)
	root, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}

	// String
	if v, err := root.Get("table.key"); err != nil {
		t.Fatal("table.key", err)
	} else {
		var str string
		err = Unmarshal(v, &str)
		if err != nil || str != "value" {
			t.Fatal("table.key", err, str)
		}
	}

	// Integer
	if v, err := root.Get("integer.key1"); err != nil {
		t.Fatal("integer.key1", err)
	} else {
		var i int
		err = Unmarshal(v, &i)
		if err != nil || i != 99 {
			t.Fatal("integer.key1", err, i)
		}
	}
	if v, err := root.Get("integer.underscores.key1"); err != nil {
		t.Fatal("integer.underscores.key1", err)
	} else {
		var i uint
		err = Unmarshal(v, &i)
		if err != nil || i != 1000 {
			t.Fatal("integer.underscores.key1", err, i)
		}
	}

	// Float
	if v, err := root.Get("float.fractional.key2"); err != nil {
		t.Fatal("float.fractional.key2", err)
	} else {
		var f float64
		err = Unmarshal(v, &f)
		if err != nil || f-3.1415 > 1e-5 {
			t.Fatal("float.fractional.key2", err, f)
		}
	}

	// Boolean
	if v, err := root.Get("boolean.True"); err != nil {
		t.Fatal("boolean.True", err)
	} else {
		var b bool
		err = Unmarshal(v, &b)
		if err != nil || !b {
			t.Fatal("boolean.True", err, b)
		}
	}

	// Datetime
	if v, err := root.Get("datetime.key3"); err != nil {
		t.Fatal("datetime.key3", err)
	} else {
		var tt time.Time
		err = Unmarshal(v, &tt)
		if err != nil || tt.Format(time.RFC3339Nano) != "1979-05-27T00:32:00.999999-07:00" {
			t.Fatal("datetime.key3", err, tt)
		}
	}

	// Array
	if v, err := root.Get("array.key1"); err != nil {
		t.Fatal("array.key1", err)
	} else {
		var arr []int
		err = Unmarshal(v, &arr)
		if err != nil || arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
			t.Fatal("array.key1", err, arr)
		}
	}
	if v, err := root.Get("array.key2"); err != nil {
		t.Fatal("array.key2", err)
	} else {
		var arr []string
		err = Unmarshal(v, &arr)
		if err != nil || arr[0] != "red" || arr[1] != "yellow" || arr[2] != "green" {
			t.Fatal("array.key2", err, arr)
		}
	}
	if v, err := root.Get("array.key3"); err != nil {
		t.Fatal("array.key3", err)
	} else {
		var arr [][]int
		err = Unmarshal(v, &arr)
		if err != nil || arr[0][0] != 1 || arr[0][1] != 2 || arr[1][0] != 3 || arr[1][1] != 4 || arr[1][2] != 5 {
			t.Fatal("array.key3", err, arr)
		}
	}

	// Map
	if v, err := root.Get("integer.underscores"); err != nil {
		t.Fatal("integer.underscores", err)
	} else {
		var m map[string]int
		err = Unmarshal(v, &m)
		if err != nil || m["key1"] != 1000 || m["key2"] != 5349221 || m["key3"] != 12345 {
			t.Fatal("integer.underscores", err, m)
		}
	}
	if v, err := root.Get("float"); err != nil {
		t.Fatal("float", err)
	} else {
		var m map[string]map[string]float64
		err = Unmarshal(v, &m)
		if err != nil || m["fractional"]["key1"]-1 > 1e-5 || m["fractional"]["key2"]-3.1415 > 1e-5 ||
			m["fractional"]["key3"] - -0.01 > 1e-5 || m["exponent"]["key1"]-5e22 > 1e-5 ||
			m["exponent"]["key2"]-1e6 > 1e-5 || m["exponent"]["key3"] - -2E-2 > 1e-5 ||
			m["both"]["key"] != 6.626e-34 || m["underscores"]["key1"]-9224617.445991228313 > 1e-5 {
			t.Fatal("integer.underscores", err, m)
		}
	}

	// Struct
	if v, err := root.Get("table.inline"); err != nil {
		t.Fatal("table.inline", err)
	} else {
		var s struct {
			Name struct {
				First string `config:"first"`
				Last  string `config:"last"`
			} `config:"name"`
			Point struct {
				X int `config:"x"`
				Y int `config:"y"`
			} `config:"point"`
		}
		err = Unmarshal(v, &s)
		if err != nil || s.Name.First != "Tom" || s.Name.Last != "Preston-Werner" || s.Point.X != 1 || s.Point.Y != 2 {
			t.Fatal("table.inline", err, s)
		}
	}

	// Array of structs
	if v, err := root.Get("products"); err != nil {
		t.Fatal("products", err)
	} else {
		var arr []*struct {
			Name  string `config:"name"`
			Sku   int    `config:"sku"`
			Color string `config:"color"`
		}
		err = Unmarshal(v, &arr)
		if err != nil || arr[0].Name != "Hammer" || arr[0].Sku != 738594937 ||
			arr[2].Name != "Nail" || arr[2].Sku != 284758393 || arr[2].Color != "gray" {
			t.Fatal("products", err, arr)
		}
	}

	// Complex one
	if v, err := root.Get("fruit"); err != nil {
		t.Fatal("fruit", err)
	} else {
		var arr []struct {
			Name     string `config:"name"`
			Physical struct {
				Color string `config:"color"`
				Shape string `config:"shape"`
			} `config:"physical"`
			Variety []struct {
				Name string `config:"name"`
			} `config:"variety"`
		}
		err = Unmarshal(v, &arr)
		if err != nil || arr[0].Name != "apple" || arr[0].Physical.Color != "red" || arr[0].Physical.Shape != "round" ||
			arr[0].Variety[0].Name != "red delicious" || arr[0].Variety[1].Name != "granny smith" ||
			arr[1].Name != "banana" || arr[1].Variety[0].Name != "plantain" {
			t.Fatal("products", err, arr)
		}
	}
}
