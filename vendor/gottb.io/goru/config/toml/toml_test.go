package toml

import (
	"os"
	"testing"
	"time"

	"gottb.io/goru/errors"
)

func TestToml(t *testing.T) {
	r, err := os.Open("example-v0.4.0.toml")
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(r)
	root, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}

	// Table
	table, err := root.Get("table")
	if err != nil || table.Len() != 3 {
		t.Fatal("table", err)
	}
	if _, err = root.Get("not found"); !errors.Is(err, ErrNotFound) {
		t.Fatal("not found", err)
	}
	tableKeys := table.Keys()
	if len(tableKeys) != 3 {
		t.Fatal("table key", table.Keys())
	}
	if v, err := table.Get("key"); err != nil || v.(*Value).value.(string) != "value" {
		t.Fatal("table.key", v, err)
	}
	if v, err := table.Get("subtable.key"); err != nil || v.(*Value).value.(string) != "another value" {
		t.Fatal("table.subtable.key", v, err)
	}
	if v, err := root.Get("x.y.z.w"); err != nil || v.(*Value).objType != TomlTable {
		t.Fatal("x.y.z.w", v, err)
	}

	// Inline table
	inlineTable, err := root.Get("table.inline")
	if err != nil {
		t.Fatal("table.inline", err)
	}
	if v, err := inlineTable.Get("name.first"); err != nil || v.(*Value).value.(string) != "Tom" {
		t.Fatal("table.inline.name.first", v, err)
	}
	if v, err := inlineTable.Get("name.last"); err != nil || v.(*Value).value.(string) != "Preston-Werner" {
		t.Fatal("table.inline.name.last", v, err)
	}
	if v, err := inlineTable.Get("point.x"); err != nil || v.(*Value).value.(int64) != 1 {
		t.Fatal("table.inline.point.x", v, err)
	}
	if v, err := inlineTable.Get("point.y"); err != nil || v.(*Value).value.(int64) != 2 {
		t.Fatal("table.inline.point.y", v, err)
	}

	// String
	if v, err := root.Get("string.basic.basic"); err != nil || v.(*Value).value.(string) != "I'm a string. \"You can quote me\". Name\tJos\u00E9\nLocation\tSF. Con chó kêu ẳng ẳng đá con mèo. 私は \u554a" {
		t.Fatal("string.basic.basic", v, err)
	}
	for _, k := range []string{"key1", "key2", "key3"} {
		if v, err := root.Get("string.multiline." + k); err != nil || v.(*Value).value.(string) != "One\nTwo" {
			t.Fatal("string.multiline."+k, v, err)
		}
	}
	for _, k := range []string{"key1", "key2", "key3"} {
		if v, err := root.Get("string.multiline.continued." + k); err != nil || v.(*Value).value.(string) != "The quick brown fox jumps over the lazy dog." {
			t.Fatal("string.multiline."+k, v, err)
		}
	}
	if v, err := root.Get("string.literal.winpath"); err != nil || v.(*Value).value.(string) != "C:\\Users\\nodejs\\templates" {
		t.Fatal("string.literal.winpath", v, err)
	}
	if v, err := root.Get("string.literal.winpath2"); err != nil || v.(*Value).value.(string) != "\\\\ServerX\\admin$\\system32\\" {
		t.Fatal("string.literal.winpath2", v, err)
	}
	if v, err := root.Get("string.literal.quoted"); err != nil || v.(*Value).value.(string) != "Tom \"Dubs\" Preston-Werner" {
		t.Fatal("string.literal.quoted", v, err)
	}
	if v, err := root.Get("string.literal.regex"); err != nil || v.(*Value).value.(string) != "<\\i\\c*\\s*>" {
		t.Fatal("string.literal.regex", v, err)
	}
	if v, err := root.Get("string.literal.multiline.regex2"); err != nil || v.(*Value).value.(string) != "I [dw]on't need \\d{2} apples" {
		t.Fatal("string.literal.multiline.regex2", v, err)
	}
	if v, err := root.Get("string.literal.multiline.lines"); err != nil || v.(*Value).value.(string) != "The first newline is\ntrimmed in raw strings.\n   All other whitespace\n   is preserved.\n" {
		t.Fatal("string.literal.multiline.lines", v, err)
	}

	// Integer
	if v, err := root.Get("integer.key1"); err != nil || v.(*Value).value.(int64) != 99 {
		t.Fatal("integer.key1", v, err)
	}
	if v, err := root.Get("integer.key2"); err != nil || v.(*Value).value.(int64) != 42 {
		t.Fatal("integer.key2", v, err)
	}
	if v, err := root.Get("integer.key3"); err != nil || v.(*Value).value.(int64) != 0 {
		t.Fatal("integer.key3", v, err)
	}
	if v, err := root.Get("integer.key4"); err != nil || v.(*Value).value.(int64) != -17 {
		t.Fatal("integer.key4", v, err)
	}
	if v, err := root.Get("integer.underscores.key1"); err != nil || v.(*Value).value.(int64) != 1000 {
		t.Fatal("integer.underscores.key1", v, err)
	}
	if v, err := root.Get("integer.underscores.key2"); err != nil || v.(*Value).value.(int64) != 5349221 {
		t.Fatal("integer.underscores.key2", v, err)
	}
	if v, err := root.Get("integer.underscores.key3"); err != nil || v.(*Value).value.(int64) != 12345 {
		t.Fatal("integer.underscores.key3", v, err)
	}

	// Float
	if v, err := root.Get("float.fractional.key1"); err != nil || v.(*Value).value.(float64)-1 > 1e-5 {
		t.Fatal("float.fractional.key1", v, err)
	}
	if v, err := root.Get("float.fractional.key2"); err != nil || v.(*Value).value.(float64)-3.1415 > 1e-5 {
		t.Fatal("float.fractional.key2", v, err)
	}
	if v, err := root.Get("float.fractional.key3"); err != nil || v.(*Value).value.(float64) - -0.01 > 1e-5 {
		t.Fatal("float.fractional.key3", v, err)
	}
	if v, err := root.Get("float.exponent.key1"); err != nil || v.(*Value).value.(float64)-5e22 > 1e-5 {
		t.Fatal("float.exponent.key1", v, err)
	}
	if v, err := root.Get("float.exponent.key2"); err != nil || v.(*Value).value.(float64)-1e6 > 1e-5 {
		t.Fatal("float.exponent.key2", v, err)
	}
	if v, err := root.Get("float.exponent.key3"); err != nil || v.(*Value).value.(float64) - -2e-2 > 1e-5 {
		t.Fatal("float.exponent.key3", v, err)
	}
	if v, err := root.Get("float.underscores.key1"); err != nil || v.(*Value).value.(float64)-9224617.445991228313 > 1e-5 {
		t.Fatal("float.underscores.key1", v, err)
	}

	// Boolean
	if v, err := root.Get("boolean.True"); err != nil || !v.(*Value).value.(bool) {
		t.Fatal("boolean.True", v, err)
	}
	if v, err := root.Get("boolean.False"); err != nil || v.(*Value).value.(bool) {
		t.Fatal("boolean.False", v, err)
	}

	// Datetime
	if v, err := root.Get("datetime.key1"); err != nil || v.(*Value).value.(time.Time).Format(time.RFC3339Nano) != "1979-05-27T07:32:00Z" {
		t.Fatal("datetime.key1", v, err)
	}
	if v, err := root.Get("datetime.key2"); err != nil || v.(*Value).value.(time.Time).Format(time.RFC3339Nano) != "1979-05-27T00:32:00-07:00" {
		t.Fatal("datetime.key2", v, err)
	}
	if v, err := root.Get("datetime.key3"); err != nil || v.(*Value).value.(time.Time).Format(time.RFC3339Nano) != "1979-05-27T00:32:00.999999-07:00" {
		t.Fatal("datetime.key3", v, err)
	}

	// Array
	array1, err := root.Get("array.key1")
	if err != nil || array1.Len() != 3 {
		t.Fatal("array.key1", err)
	}
	if v, err := array1.At(0); err != nil || v.(*Value).value.(int64) != 1 {
		t.Fatal("array.key1.0", v, err)
	}
	if v, err := array1.At(1); err != nil || v.(*Value).value.(int64) != 2 {
		t.Fatal("array.key1.1", v, err)
	}
	if v, err := array1.At(2); err != nil || v.(*Value).value.(int64) != 3 {
		t.Fatal("array.key1.2", v, err)
	}
	if v, err := root.Get("array.key2.0"); err != nil || v.(*Value).value.(string) != "red" {
		t.Fatal("array.key2.0", v, err)
	}
	if v, err := root.Get("array.key2.1"); err != nil || v.(*Value).value.(string) != "yellow" {
		t.Fatal("array.key2.1", v, err)
	}
	if v, err := root.Get("array.key2.2"); err != nil || v.(*Value).value.(string) != "green" {
		t.Fatal("array.key2.2", v, err)
	}
	if v, err := root.Get("array.key3.0.0"); err != nil || v.(*Value).value.(int64) != 1 {
		t.Fatal("array.key3.0.0", v, err)
	}
	if v, err := root.Get("array.key3.0.1"); err != nil || v.(*Value).value.(int64) != 2 {
		t.Fatal("array.key3.0.1", v, err)
	}
	if v, err := root.Get("array.key3.1.0"); err != nil || v.(*Value).value.(int64) != 3 {
		t.Fatal("array.key3.1.0", v, err)
	}
	if v, err := root.Get("array.key3.1.1"); err != nil || v.(*Value).value.(int64) != 4 {
		t.Fatal("array.key3.1.1", v, err)
	}
	if v, err := root.Get("array.key3.1.2"); err != nil || v.(*Value).value.(int64) != 5 {
		t.Fatal("array.key3.1.2", v, err)
	}
	if v, err := root.Get("array.key4.0.0"); err != nil || v.(*Value).value.(int64) != 1 {
		t.Fatal("array.key4.0.0", v, err)
	}
	if v, err := root.Get("array.key4.0.1"); err != nil || v.(*Value).value.(int64) != 2 {
		t.Fatal("array.key4.0.1", v, err)
	}
	if v, err := root.Get("array.key4.1.0"); err != nil || v.(*Value).value.(string) != "a" {
		t.Fatal("array.key4.1.0", v, err)
	}
	if v, err := root.Get("array.key4.1.1"); err != nil || v.(*Value).value.(string) != "b" {
		t.Fatal("array.key4.1.1", v, err)
	}
	if v, err := root.Get("array.key4.1.2"); err != nil || v.(*Value).value.(string) != "c" {
		t.Fatal("array.key4.1.2", v, err)
	}
	if v, err := root.Get("array.key5.0"); err != nil || v.(*Value).value.(int64) != 1 {
		t.Fatal("array.key5.0", v, err)
	}
	if v, err := root.Get("array.key5.1"); err != nil || v.(*Value).value.(int64) != 2 {
		t.Fatal("array.key5.1", v, err)
	}
	if v, err := root.Get("array.key5.2"); err != nil || v.(*Value).value.(int64) != 3 {
		t.Fatal("array.key5.2", v, err)
	}
	if v, err := root.Get("array.key6.0"); err != nil || v.(*Value).value.(int64) != 1 {
		t.Fatal("array.key6.0", v, err)
	}
	if v, err := root.Get("array.key6.1"); err != nil || v.(*Value).value.(int64) != 2 {
		t.Fatal("array.key6.1", v, err)
	}

	// Array of tables
	products, err := root.Get("products")
	if err != nil || products.Len() != 3 {
		t.Fatal("products", err)
	}
	product1, err := products.At(0)
	if err != nil {
		t.Fatal("products.0", err)
	}
	if v, err := product1.Get("name"); err != nil || v.(*Value).value.(string) != "Hammer" {
		t.Fatal("products.0.name", v, err)
	}
	if v, err := product1.Get("sku"); err != nil || v.(*Value).value.(int64) != 738594937 {
		t.Fatal("products.0.sku", v, err)
	}
	product3, err := products.At(2)
	if err != nil {
		t.Fatal("products.2", err)
	}
	if v, err := product3.Get("name"); err != nil || v.(*Value).value.(string) != "Nail" {
		t.Fatal("products.2.name", v, err)
	}
	if v, err := product3.Get("sku"); err != nil || v.(*Value).value.(int64) != 284758393 {
		t.Fatal("products.2.sku", v, err)
	}
	if v, err := product3.Get("color"); err != nil || v.(*Value).value.(string) != "gray" {
		t.Fatal("products.2.color", v, err)
	}
	if v, err := root.Get("fruit.0.name"); err != nil || v.(*Value).value.(string) != "apple" {
		t.Fatal("fruit.0.name", v, err)
	}
	if v, err := root.Get("fruit.0.physical.color"); err != nil || v.(*Value).value.(string) != "red" {
		t.Fatal("fruit.0.physical.color", v, err)
	}
	if v, err := root.Get("fruit.0.physical.shape"); err != nil || v.(*Value).value.(string) != "round" {
		t.Fatal("fruit.0.physical.shape", v, err)
	}
	if v, err := root.Get("fruit.0.variety.0.name"); err != nil || v.(*Value).value.(string) != "red delicious" {
		t.Fatal("fruit.0.variety.0.name", v, err)
	}
	if v, err := root.Get("fruit.0.variety.1.name"); err != nil || v.(*Value).value.(string) != "granny smith" {
		t.Fatal("fruit.0.variety.1.name", v, err)
	}
	if v, err := root.Get("fruit.1.name"); err != nil || v.(*Value).value.(string) != "banana" {
		t.Fatal("fruit.1.name", v, err)
	}
	if v, err := root.Get("fruit.1.variety.0.name"); err != nil || v.(*Value).value.(string) != "plantain" {
		t.Fatal("fruit.1.variety.0.name", v, err)
	}
}
