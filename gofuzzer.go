package gofuzzer

import (
	"encoding/binary"
	"fmt"
	"github.com/ret2happy/gofuzzer/utils"
	"io/ioutil"
	"math/big"
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

func ConsumeBool(cursor *int, fuzzData []byte) (bool, error) {
	bakCursor := *cursor
	*cursor += 1
	if bakCursor >= len(fuzzData) {
		return false, fmt.Errorf("exhausted all fuzz data")
	}
	return fuzzData[bakCursor]%2 == 0, nil
}

func ConsumeUint8Range(cursor *int, fuzzData []byte, maxNum uint8) (uint8, error) {
	bakCursor := *cursor
	*cursor += 1
	if bakCursor >= len(fuzzData) {
		*cursor = len(fuzzData)
		return 0, fmt.Errorf("exhausted all fuzz data")
	}
	return fuzzData[bakCursor] % maxNum, nil
}

func ConsumeUint64(cursor *int, fuzzData []byte) (uint64, error) {
	bakCursor := *cursor
	byteLength := 8
	if bakCursor >= len(fuzzData) || len(fuzzData)-bakCursor < byteLength {
		*cursor = len(fuzzData)
		return 0, fmt.Errorf("exhausted all fuzz data")
	}
	*cursor += byteLength
	return binary.LittleEndian.Uint64(fuzzData[bakCursor : bakCursor+byteLength]), nil
}

func ConsumeInt64(cursor *int, fuzzData []byte) (int64, error) {
	result, err := ConsumeUint64(cursor, fuzzData)
	return int64(result), err
}

func ConsumeBigInt(cursor *int, fuzzData []byte, bitLength uint) (big.Int, error) {
	bakCursor := *cursor
	targetByteLength := int(bitLength / 8)
	var result big.Int
	if bakCursor >= len(fuzzData) || bakCursor+targetByteLength > len(fuzzData) || targetByteLength > 32 {
		*cursor = len(fuzzData)
		return result, fmt.Errorf("exhausted all fuzz data")
	}
	*cursor += targetByteLength
	result.SetBytes(fuzzData[bakCursor : bakCursor+targetByteLength])
	return result, nil
}

func ConsumeArray[K any](cursor *int, fuzzData []byte, candidates []K) (K, error) {
	offset, err := ConsumeUint8Range(cursor, fuzzData, uint8(len(candidates)))
	if err != nil {
		return nil, err
	}
	return candidates[offset], nil
}
