package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func BenchmarkGetDomainStat(b *testing.B) {
	r, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer r.Close()

	dataFile, err := r.File[0].Open()
	if err != nil {
		b.Fatal(err)
	}

	data, err := io.ReadAll(dataFile)
	if err != nil {
		dataFile.Close()
		b.Fatal(err)
	}

	if err := dataFile.Close(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, err := GetDomainStat(bytes.NewReader(data), "biz")
		if err != nil {
			b.Fatal(err)
		}
	}
}
