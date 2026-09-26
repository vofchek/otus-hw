package main

import (
	"flag"
	"log"
	"os"

	"github.com/cheggaaa/pb"
)

var (
	from, to      string
	limit, offset int64
)

// размер увеличения лимита копирования.
var copyLimit int64 = 500

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()

	fromStat, err := os.Stat(from)
	if err != nil {
		log.Fatalf("cant open from %v", err)
	}

	if fromStat.Size() < offset {
		log.Fatalf("offset must be lesser or equal to file size")
	}

	var maxCopy int64
	if limit == 0 {
		maxCopy = fromStat.Size() - offset
	} else {
		maxCopy = limit
	}

	steps := maxCopy / copyLimit

	if steps*copyLimit < maxCopy {
		steps++
	}

	progressBar := pb.New(int(steps))
	progressBar.ShowCounters = true

	cLimit := copyLimit
	if limit > 0 && limit < copyLimit {
		cLimit = limit
	}

	// [Copy()] полностью перезаписывает файл,
	// поэтому на каждой итерации заново перезаписываем файл всё большей порцией данных из исходного файла.
	// Порционное копирование сделано специально для реализации прогресс бара.
	for i := int64(0); i < steps; i++ {
		copyErr := Copy(from, to, offset, cLimit+copyLimit*i)

		if copyErr != nil {
			log.Fatal(copyErr)
		}

		progressBar.Increment()
	}

	progressBar.FinishPrint("Copy done!")
}
