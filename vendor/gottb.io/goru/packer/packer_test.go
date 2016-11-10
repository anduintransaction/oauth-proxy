package packer

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func testPack(t *testing.T, pack *Pack) {
	originalContent, err := ioutil.ReadFile("template/main.go.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	f, err := pack.Open("main.go.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	packedContent, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(originalContent, packedContent) != 0 {
		t.Fatal("content not matched")
	}
	f.Close()
	_, err = ioutil.ReadAll(f)
	if err == nil {
		t.Fatal("read after closed")
	}
	f, err = pack.Open("not found")
	if err == nil {
		t.Fatal("not found")
	}
}

func TestPacker(t *testing.T) {
	testPack(t, VVVPack)
}

func TestSeek(t *testing.T) {
	originalFile, err := os.Open("template/main.go.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	defer originalFile.Close()
	originalFile.Seek(100, 0)
	originalContent, err := ioutil.ReadAll(originalFile)
	if err != nil {
		t.Fatal(err)
	}
	packedFile, err := VVVPack.Open("main.go.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	packedFile.Seek(100, 0)
	packedContent, err := ioutil.ReadAll(packedFile)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(originalContent, packedContent) != 0 {
		t.Fatal("seeked content not matched")
	}
}

func TestPackerDevMode(t *testing.T) {
	pack := &Pack{
		root: "template",
		dev:  true,
	}
	testPack(t, pack)
}
