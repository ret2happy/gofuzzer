package gofuzzer

import "testing"

func TestConsumeArray(t *testing.T) {
	var tests = []struct {
		desc       string
		fuzzData   []byte
		cursor     int
		candidates []int
		want       int
		reject     bool
	}{
		{
			desc:       "choose from int array normally",
			fuzzData:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			cursor:     0,
			candidates: []int{1, 2, 3, 4, 5},
			want:       2,
		},
		{
			desc:       "choose from int array normally with the edge bound",
			fuzzData:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			cursor:     7,
			candidates: []int{1, 2, 3, 4, 5},
			want:       4,
		},
		{
			desc:       "choose from empty int array",
			fuzzData:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			cursor:     0,
			candidates: []int{},
			reject:     true,
		},
		{
			desc:       "choose from int array with out of bound cursor",
			fuzzData:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
			cursor:     8,
			candidates: []int{1, 2, 3, 4, 5},
			reject:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result, err := ConsumeArray(&test.cursor, test.fuzzData, test.candidates)
			if test.reject && err == nil {
				t.Fatalf("Expect error in ConsumeArray")
			}
			if !test.reject && test.want != result {
				t.Fatalf("Wrong consume result in ConsumeArray. Got %d, expect: %d", result, test.want)
			}
		})
	}
}
