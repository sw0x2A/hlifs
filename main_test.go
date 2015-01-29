package hlifs

import (
	"bytes"
	"testing"
)

func TestHashFile(t *testing.T) {
	files := [...]FileData{
		{name: "test/datafile1", hash: []byte{178, 129, 202, 107, 136, 158, 162, 98, 129, 252, 24, 214, 195, 171, 38, 33}},
		{name: "test/textfile1", hash: []byte{67, 189, 211, 69, 178, 12, 210, 76, 57, 147, 48, 77, 165, 188, 7, 112}},
	}
	for _, file := range files {
		in, out := file.name, file.hash
		if x, err := HashFile(in); bytes.Compare(x, out) != 0 {
			if err != nil {
				t.Errorf("Hashfile(%s) = %d returned err %s", in, x, err)
			} else {
				t.Errorf("Hashfile(%s) = %d, should be %d", in, x, out)
			}
		}
	}
}

func BenchmarkHashFile(b *testing.B) {
	in := "test/datafile1"
	for i := 0; i < b.N; i++ {
		HashFile(in)
	}
}
