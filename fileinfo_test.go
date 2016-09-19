package ddet

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestMd5(t *testing.T) {
	file, _ := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())

	ioutil.WriteFile(file.Name(), []byte("constant text string"), 0644)

	md5, _ := ComputeMd5(file.Name())
	t.Log("Path=", file.Name(), ", MD5=", md5)

	expected := []byte{196, 84, 116, 50, 184, 145, 163, 237, 46, 107, 22, 229, 141, 66, 249, 123}
	if !bytes.Equal(md5, expected) {
		t.Error("bad md5=", md5, ", expected=", expected)
	}
}

func TestSize(t *testing.T) {
	file, _ := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())

	ioutil.WriteFile(file.Name(), []byte("constant text string"), 0644)

	size, _, _ := GetFileStats(file.Name())

	expected := int64(20)
	if size != expected {
		t.Error("bad size=", size, ", expected=", expected)
	}
}
