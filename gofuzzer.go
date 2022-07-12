package gofuzzer

import (
	"fmt"
	"github.com/ret2happy/gofuzzer/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func UnmarshalCorpusFile(b []byte) ([]any, error) {
	return utils.UnmarshalCorpusFile(b)
}

type FuzzCoreFunc func(corpus []any) (err error)

func DumpFuzzCoreCoverage(t *testing.T, corpusPath string, callback FuzzCoreFunc) {
	files, err := ioutil.ReadDir(corpusPath)
	if err != nil {
		return
	}
	fmt.Printf("Collected %d corpus files.\n", len(files))
	for idx, f := range files {
		fmt.Printf("Processing %d/%d\r", idx, len(files))
		filePath := filepath.Join(corpusPath, f.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		originData, err := UnmarshalCorpusFile(data)
		if err != nil {
			t.Fatal(err)
		}
		err = callback(originData)
		if err != nil {
			t.Fatal(err)
		}
	}
	println("All corpus files processed.")
}
