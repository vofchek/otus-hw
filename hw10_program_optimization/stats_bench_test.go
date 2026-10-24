package hw10programoptimization_test

import (
	"archive/zip"
	"log"
	"testing"

	hw10programoptimization "github.com/vofchek/otus-hw/hw10_program_optimization"
)

func BenchmarkGetDomainStat(b *testing.B) {
	r, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	data, err := r.File[0].Open()

	for i := 0; i < b.N; i++ {
		hw10programoptimization.GetDomainStat(data, "biz")
	}
}
