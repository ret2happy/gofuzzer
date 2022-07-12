package gofuzzer

import (
	"github.com/ret2happy/gofuzzer/utils"
)

func UnmarshalCorpusFile(b []byte) ([]any, error) {
	vals, err := utils.UnmarshalCorpusFile(b)
	return vals, err
}
