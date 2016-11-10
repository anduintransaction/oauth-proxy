package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPkgPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path, err := PkgPath(cwd)
	if err != nil {
		t.Fatal(err)
	}
	if path != "gottb.io/goru/utils" {
		t.Fatal("package path: ", path)
	}
}

func TestWalk(t *testing.T) {
	root, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	dirs := []string{
		"1/1.1",
		"1/1.2",
		"2/2.1",
		"3",
		"4/4.1",
		"4/4.2",
		"5/5.1",
	}
	files := []string{
		"1/1.1/1.1.1.dat",
		"1/1.1/1.1.2.dat",
		"1/1.2/1.2.1.dat",
		"2/2.2.dat",
		"3/3.1.dat",
		"4/4.2/4.2.1.dat",
		"4/4.2/4.2.2.dat",
	}
	ignores := map[string]string{
		"ignore":       "5\n",
		"4/ignore":     "4.1\n",
		"4/4.2/ignore": "*2.2*\n",
	}
	for _, dir := range dirs {
		err = os.MkdirAll(filepath.Join(root, dir), 0755)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range files {
		err = ioutil.WriteFile(filepath.Join(root, file), []byte{}, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	for file, content := range ignores {
		err = ioutil.WriteFile(filepath.Join(root, file), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	tree := map[string]string{
		"":                "",
		"1":               "",
		"1/1.1":           "",
		"1/1.1/1.1.1.dat": "",
		"1/1.1/1.1.2.dat": "",
		"1/1.2":           "",
		"1/1.2/1.2.1.dat": "",
		"2":               "",
		"2/2.1":           "",
		"2/2.2.dat":       "",
		"3":               "",
		"3/3.1.dat":       "",
		"4":               "",
		"4/4.2":           "",
		"4/4.2/4.2.1.dat": "",
	}
	err = Walk(root, func(path string, info os.FileInfo) error {
		relativePath := strings.TrimPrefix(filepath.ToSlash(strings.TrimPrefix(path, root)), "/")
		_, ok := tree[relativePath]
		if !ok {
			t.Fatal("not exist in tree: ", relativePath)
		}
		delete(tree, relativePath)
		return nil
	}, "ignore")
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 0 {
		t.Fatal("not walked: ", tree)
	}
}
