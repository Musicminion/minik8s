package zip

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// 删除./testExample/example1-output/example1.zip 如果有的话
	os.RemoveAll("./testExample/example1-output/")
	os.Mkdir("./testExample/example1-output/", os.ModePerm)
	// 删除./testExample/example2-output/example2.zip 如果有的话
	os.RemoveAll("./testExample/example2-output/")
	os.Mkdir("./testExample/example2-output/", os.ModePerm)
	m.Run()
}

func TestZip(t *testing.T) {
	// TODO
	err := CompressToZip("./testExample/example1", "./testExample/example1-output/example1.zip")

	if err != nil {
		t.Log(err)
	}

	err = CompressToZip("./testExample/example2", "./testExample/example2-output/example2.zip")

	if err != nil {
		t.Log(err)
	}
}

func TestUnzip(t *testing.T) {
	// TODO
	err := DecompressZip("./testExample/example1-output/example1.zip", "./testExample/example1-output/")

	if err != nil {
		t.Log(err)
	}

	err = DecompressZip("./testExample/example2-output/example2.zip", "./testExample/example2-output/")

	if err != nil {
		t.Log(err)
	}
}

func TestClearAll(t *testing.T) {
	// 删除./testExample/example1-output/example1.zip 如果有的话
	os.RemoveAll("./testExample/example1-output/")
	os.Mkdir("./testExample/example1-output/", os.ModePerm)
	// 删除./testExample/example2-output/example2.zip 如果有的话
	os.RemoveAll("./testExample/example2-output/")
	os.Mkdir("./testExample/example2-output/", os.ModePerm)
}
