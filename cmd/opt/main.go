package main

import (
	"os"

	"github.com/romshark/jpegbench/optimized/jpeg" // Optimized.
)

func main() {
	f, err := os.OpenFile("testdata/600x239_35kb.jpg", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	i, err := jpeg.Decode(f)
	if err != nil {
		panic(err)
	}
	b := i.Bounds()
	print(b.Dx(), "x", b.Dy())
}
