package main

import (
	"bytes"
	"image/jpeg"
	"os/exec"
	"testing"

	optjpeg "github.com/romshark/jpegbench/optimized/jpeg"
	"github.com/stretchr/testify/require"
)

func TestDecodeEncodeCompare(t *testing.T) {
	require.DirExists(t, "testdata")
	for file, contents := range filesJPEG(t, "testdata") {
		t.Run(file, func(t *testing.T) {
			r := bytes.NewReader(contents)

			r.Reset(contents)
			i1, err := jpeg.Decode(r)
			require.NoError(t, err)
			require.NotZero(t, i1)

			r.Reset(contents)
			i2, err := optjpeg.Decode(r)
			require.NoError(t, err)
			require.NotZero(t, i2)

			encOptions := &jpeg.Options{
				Quality: 100,
			}

			var b1 bytes.Buffer
			var b2 bytes.Buffer

			err = jpeg.Encode(&b1, i1, encOptions)
			require.NoError(t, err)
			err = jpeg.Encode(&b2, i2, encOptions)
			require.NoError(t, err)

			require.Equal(t, b1.Bytes(), b2.Bytes())
		})
	}
}

func TestExecutables(t *testing.T) {
	outStd, err := exec.Command("go", "run", "./cmd/std").CombinedOutput()
	if err != nil {
		t.Fatalf("failed executing cmd/std: %v\nOutput: %s", err, string(outStd))
	}

	outOpt, err := exec.Command("go", "run", "./cmd/opt").CombinedOutput()
	if err != nil {
		t.Fatalf("failed executing cmd/opt: %v\nOutput: %s", err, string(outOpt))
	}

	require.Equal(t, outStd, outOpt)
}
