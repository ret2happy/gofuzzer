package utils

import (
	"strconv"
	"testing"
)

func TestUnmarshalMarshal(t *testing.T) {
	var tests = []struct {
		desc   string
		in     string
		reject bool
		want   string // if different from in
	}{
		{
			desc:   "missing version",
			in:     "int(1234)",
			reject: true,
		},
		{
			desc: "malformed string",
			in: `go test fuzz v1
string("a"bcad")`,
			reject: true,
		},
		{
			desc: "empty value",
			in: `go test fuzz v1
int()`,
			reject: true,
		},
		{
			desc: "negative uint",
			in: `go test fuzz v1
uint(-32)`,
			reject: true,
		},
		{
			desc: "int8 too large",
			in: `go test fuzz v1
int8(1234456)`,
			reject: true,
		},
		{
			desc: "multiplication in int value",
			in: `go test fuzz v1
int(20*5)`,
			reject: true,
		},
		{
			desc: "double negation",
			in: `go test fuzz v1
int(--5)`,
			reject: true,
		},
		{
			desc: "malformed bool",
			in: `go test fuzz v1
bool(0)`,
			reject: true,
		},
		{
			desc: "malformed byte",
			in: `go test fuzz v1
byte('aa)`,
			reject: true,
		},
		{
			desc: "byte out of range",
			in: `go test fuzz v1
byte('☃')`,
			reject: true,
		},
		{
			desc: "extra newline",
			in: `go test fuzz v1
string("has extra newline")
`,
			want: `go test fuzz v1
string("has extra newline")`,
		},
		{
			desc: "trailing spaces",
			in: `go test fuzz v1
string("extra")
[]byte("spacing")  
    `,
			want: `go test fuzz v1
string("extra")
[]byte("spacing")`,
		},
		{
			desc: "float types",
			in: `go test fuzz v1
float64(0)
float32(0)`,
		},
		{
			desc: "various types",
			in: `go test fuzz v1
int(-23)
int8(-2)
int64(2342425)
uint(1)
uint16(234)
uint32(352342)
uint64(123)
rune('œ')
byte('K')
byte('ÿ')
[]byte("hello¿")
[]byte("a")
bool(true)
string("hello\\xbd\\xb2=\\xbc ⌘")
float64(-12.5)
float32(2.5)`,
		},
		{
			desc: "float edge cases",
			// The two IEEE 754 bit patterns used for the math.Float{64,32}frombits
			// encodings are non-math.NAN quiet-NaN values. Since they are not equal
			// to math.NaN(), they should be re-encoded to their bit patterns. They
			// are, respectively:
			//   * math.Float64bits(math.NaN())+1
			//   * math.Float32bits(float32(math.NaN()))+1
			in: `go test fuzz v1
float32(-0)
float64(-0)
float32(+Inf)
float32(-Inf)
float32(NaN)
float64(+Inf)
float64(-Inf)
float64(NaN)
math.Float64frombits(0x7ff8000000000002)
math.Float32frombits(0x7fc00001)`,
		},
		{
			desc: "int variations",
			// Although we arbitrarily choose default integer bases (0 or 16), we may
			// want to change those arbitrary choices in the future and should not
			// break the parser. Verify that integers in the opposite bases still
			// parse correctly.
			in: `go test fuzz v1
int(0x0)
int32(0x41)
int64(0xfffffffff)
uint32(0xcafef00d)
uint64(0xffffffffffffffff)
uint8(0b0000000)
byte(0x0)
byte('\000')
byte('\u0000')
byte('\'')
math.Float64frombits(9221120237041090562)
math.Float32frombits(2143289345)`,
			want: `go test fuzz v1
int(0)
rune('A')
int64(68719476735)
uint32(3405705229)
uint64(18446744073709551615)
byte('\x00')
byte('\x00')
byte('\x00')
byte('\x00')
byte('\'')
math.Float64frombits(0x7ff8000000000002)
math.Float32frombits(0x7fc00001)`,
		},
		{
			desc: "rune validation",
			in: `go test fuzz v1
rune(0)
rune(0x41)
rune(-1)
rune(0xfffd)
rune(0xd800)
rune(0x10ffff)
rune(0x110000)
`,
			want: `go test fuzz v1
rune('\x00')
rune('A')
int32(-1)
rune('�')
int32(55296)
rune('\U0010ffff')
int32(1114112)`,
		},
		{
			desc: "int overflow",
			in: `go test fuzz v1
int(0x7fffffffffffffff)
uint(0xffffffffffffffff)`,
			want: func() string {
				switch strconv.IntSize {
				case 32:
					return `go test fuzz v1
int(-1)
uint(4294967295)`
				case 64:
					return `go test fuzz v1
int(9223372036854775807)
uint(18446744073709551615)`
				default:
					panic("unreachable")
				}
			}(),
		},
		{
			desc: "windows new line",
			in:   "go test fuzz v1\r\nint(0)\r\n",
			want: "go test fuzz v1\nint(0)",
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			vals, err := UnmarshalCorpusFile([]byte(test.in))
			if test.reject {
				if err == nil {
					t.Fatalf("unmarshal unexpected success")
				}
				return
			}
			if err != nil {
				t.Fatalf("unmarshal unexpected error: %v", err)
			}
			newB := MarshalCorpusFile(vals...)
			if err != nil {
				t.Fatalf("marshal unexpected error: %v", err)
			}
			if newB[len(newB)-1] != '\n' {
				t.Error("didn't write final newline to corpus file")
			}

			want := test.want
			if want == "" {
				want = test.in
			}
			want += "\n"
			got := string(newB)
			if got != want {
				t.Errorf("unexpected marshaled value\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

// BenchmarkMarshalCorpusFile measures the time it takes to serialize byte
// slices of various sizes to a corpus file. The slice contains a repeating
// sequence of bytes 0-255 to mix escaped and non-escaped characters.
func BenchmarkMarshalCorpusFile(b *testing.B) {
	buf := make([]byte, 1024*1024)
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(i)
	}

	for sz := 1; sz <= len(buf); sz <<= 1 {
		sz := sz
		b.Run(strconv.Itoa(sz), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.SetBytes(int64(sz))
				MarshalCorpusFile(buf[:sz])
			}
		})
	}
}
