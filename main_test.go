package main

import (
	"bytes"
	"testing"
)

func TestFgLen(t *testing.T) {
	in := FileGroup{
		&FileData{
			name: "test/datafile1",
		},
		&FileData{
			name: "test/textfile1",
		},
	}
	out := 2
	if x := in.Len(); x != out {
		t.Errorf("fg.Len() = %d should be %d", x, out)
	}
}

func TestGetFileHash(t *testing.T) {
	files := [...]FileData{
		{
			name: "test/datafile1",
			// hash, _ := hex.DecodeString("44ecd64bb65e5ff2fdbc3c4b67f9c162b0ac41930bdba201cda7c15bde64da38")
			hash: []byte{68, 236, 214, 75, 182, 94, 95, 242, 253, 188, 60, 75, 103, 249, 193, 98, 176, 172, 65, 147, 11, 219, 162, 1, 205, 167, 193, 91, 222, 100, 218, 56},
		},
		{
			name: "test/textfile1",
			// hash, _ := hex.DecodeString("a54c4bccfd84a80fbbcb928dd98048ef33990a437157a0ac8f7999891c6bcf7d")
			hash: []byte{165, 76, 75, 204, 253, 132, 168, 15, 187, 203, 146, 141, 217, 128, 72, 239, 51, 153, 10, 67, 113, 87, 160, 172, 143, 121, 153, 137, 28, 107, 207, 125},
		},
	}
	for _, file := range files {
		in, out := file.name, file.hash
		if x, err := getFileHash(in); bytes.Compare(x, out) != 0 {
			if err != nil {
				t.Errorf("Hashfile(%s) = %d returned err %s", in, x, err)
			} else {
				t.Errorf("Hashfile(%s) = %d, should be %d", in, x, out)
			}
		}
	}
}

func BenchmarkGetFileHash(b *testing.B) {
	in := "test/datafile1"
	for i := 0; i < b.N; i++ {
		getFileHash(in)
	}
}

func TestGetRandStringBytes(t *testing.T) {
	pairs := [...][2]int{{1, 2}, {2, 4}}
	for _, pair := range pairs {
		min, max := pair[0], pair[1]
		if x, err := getRandStringBytes(min, max); len(x) < min || len(x) > max {
			if err != nil {
				t.Errorf("len(getRandStringBytes(%d, %d)) = %d (%s), returned err %s", min, max, len(x), x, err)
			} else {
				t.Errorf("len(getRandStringBytes(%d, %d)) = %d (%s), should be between %d and %d", min, max, len(x), x, min, max)
			}
		}
	}
}
