package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"testing"

	optjpeg "github.com/romshark/jpegbench/optimized/jpeg"
)

var benchFn func(io.Reader) (image.Image, error)

func init() {
	switch f := os.Getenv("BENCH_FN"); f {
	case "std", "":
		benchFn = jpeg.Decode
	case "opt":
		benchFn = optjpeg.Decode
	default:
		panic(fmt.Errorf("undefined bench fn: %q", f))
	}
}

func BenchmarkDecode(b *testing.B) {
	for filePath, contents := range filesJPEG(b, "testdata") {
		r := bytes.NewReader(contents)
		b.Run(filePath, func(b *testing.B) {
			for range b.N {
				r.Reset(contents)
				benchFn(r)
			}
		})
	}
}

func filesJPEG(tb testing.TB, dir string) iter.Seq2[string, []byte] {
	return func(yield func(string, []byte) bool) {
		err := fs.WalkDir(
			os.DirFS(dir),
			".",
			func(path string, e fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if e.IsDir() {
					return nil
				}
				ext := filepath.Ext(path)
				switch ext {
				case ".jpg", ".jpeg":
					c, err := os.ReadFile(filepath.Join(dir, path))
					if err != nil {
						tb.Fatal(err)
					}
					if !yield(e.Name(), c) {
						return errors.New("stop")
					}
				}
				return nil
			})
		if err != nil {
			tb.Fatal(err)
		}
	}
}
