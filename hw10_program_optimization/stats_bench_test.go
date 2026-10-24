package hw10programoptimization

import (
	"archive/zip"
	"log"
	"testing"
)

func BenchmarkGetDomainStat(b *testing.B) {
	r, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	data, err := r.File[0].Open()
	defer r.Close()

	if err != nil {
		r.Close()
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		GetDomainStat(data, "biz")
	}
}
