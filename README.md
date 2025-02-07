# Go's standard `image/jpeg` decoding optimization.

Official Go PR: https://github.com/golang/go/pull/71618

This repository demonstrates a potential optimization of decoding JPEG images
in the Go standard library. This is achieved by unrolling
unzig and shift-clamp loops in `scan.go`.

On average across benchmarks (Go 1.23.6), this is **~13% faster on AMD64**
and **~14%/11% faster on Apple Silicon ARM64 (M1/M4)**
at the cost of slightly larger binary size (~16KiB).

patch:
<details>

```diff
--- /usr/lib/go/src/image/jpeg/reader.go	2025-02-05 01:18:32.000000000 +0100
+++ ./optimized/jpeg/reader.go	2025-02-07 18:33:34.900861895 +0100
@@ -75,7 +76,7 @@
 // unzig maps from the zig-zag ordering to the natural ordering. For example,
 // unzig[3] is the column and row of the fourth element in zig-zag order. The
 // value is 16, which means first column (16%8 == 0) and third row (16/8 == 2).
-var unzig = [blockSize]int{
+var unzig = [blockSize]uint8{
 	0, 1, 8, 16, 9, 2, 3, 10,
 	17, 24, 32, 25, 18, 11, 4, 5,
 	12, 19, 26, 33, 40, 48, 41, 34,
diff -ruN /usr/lib/go/src/image/jpeg/scan.go ./optimized/jpeg/scan.go
--- /usr/lib/go/src/image/jpeg/scan.go	2025-02-05 01:18:32.000000000 +0100
+++ ./optimized/jpeg/scan.go	2025-02-07 22:39:23.342471482 +0100
@@ -465,9 +465,73 @@
 // to the image.
 func (d *decoder) reconstructBlock(b *block, bx, by, compIndex int) error {
 	qt := &d.quant[d.comp[compIndex].tq]
-	for zig := 0; zig < blockSize; zig++ {
-		b[unzig[zig]] *= qt[zig]
-	}
+
+	// This sequence exactly follows the indexes of the unzig mapping.
+	b[0] *= qt[0]
+	b[1] *= qt[1]
+	b[8] *= qt[2]
+	b[16] *= qt[3]
+	b[9] *= qt[4]
+	b[2] *= qt[5]
+	b[3] *= qt[6]
+	b[10] *= qt[7]
+	b[17] *= qt[8]
+	b[24] *= qt[9]
+	b[32] *= qt[10]
+	b[25] *= qt[11]
+	b[18] *= qt[12]
+	b[11] *= qt[13]
+	b[4] *= qt[14]
+	b[5] *= qt[15]
+	b[12] *= qt[16]
+	b[19] *= qt[17]
+	b[26] *= qt[18]
+	b[33] *= qt[19]
+	b[40] *= qt[20]
+	b[48] *= qt[21]
+	b[41] *= qt[22]
+	b[34] *= qt[23]
+	b[27] *= qt[24]
+	b[20] *= qt[25]
+	b[13] *= qt[26]
+	b[6] *= qt[27]
+	b[7] *= qt[28]
+	b[14] *= qt[29]
+	b[21] *= qt[30]
+	b[28] *= qt[31]
+	b[35] *= qt[32]
+	b[42] *= qt[33]
+	b[49] *= qt[34]
+	b[56] *= qt[35]
+	b[57] *= qt[36]
+	b[50] *= qt[37]
+	b[43] *= qt[38]
+	b[36] *= qt[39]
+	b[29] *= qt[40]
+	b[22] *= qt[41]
+	b[15] *= qt[42]
+	b[23] *= qt[43]
+	b[30] *= qt[44]
+	b[37] *= qt[45]
+	b[44] *= qt[46]
+	b[51] *= qt[47]
+	b[58] *= qt[48]
+	b[59] *= qt[49]
+	b[52] *= qt[50]
+	b[45] *= qt[51]
+	b[38] *= qt[52]
+	b[31] *= qt[53]
+	b[39] *= qt[54]
+	b[46] *= qt[55]
+	b[53] *= qt[56]
+	b[60] *= qt[57]
+	b[61] *= qt[58]
+	b[54] *= qt[59]
+	b[47] *= qt[60]
+	b[55] *= qt[61]
+	b[62] *= qt[62]
+	b[63] *= qt[63]
+
 	idct(b)
 	dst, stride := []byte(nil), 0
 	if d.nComp == 1 {
@@ -486,22 +550,82 @@
 			return UnsupportedError("too many components")
 		}
 	}
+
 	// Level shift by +128, clip to [0, 255], and write to dst.
-	for y := 0; y < 8; y++ {
-		y8 := y * 8
-		yStride := y * stride
-		for x := 0; x < 8; x++ {
-			c := b[y8+x]
-			if c < -128 {
-				c = 0
-			} else if c > 127 {
-				c = 255
-			} else {
-				c += 128
-			}
-			dst[yStride+x] = uint8(c)
+	writeDst := func(index int) {
+		c := (*b)[index] + 128
+		if c < 0 {
+			c = 0
+		} else if c > 255 {
+			c = 255
 		}
+		dst[(index/8)*stride+(index%8)] = uint8(c)
 	}
+	writeDst(0)
+	writeDst(1)
+	writeDst(2)
+	writeDst(3)
+	writeDst(4)
+	writeDst(5)
+	writeDst(6)
+	writeDst(7)
+	writeDst(8)
+	writeDst(9)
+	writeDst(10)
+	writeDst(11)
+	writeDst(12)
+	writeDst(13)
+	writeDst(14)
+	writeDst(15)
+	writeDst(16)
+	writeDst(17)
+	writeDst(18)
+	writeDst(19)
+	writeDst(20)
+	writeDst(21)
+	writeDst(22)
+	writeDst(23)
+	writeDst(24)
+	writeDst(25)
+	writeDst(26)
+	writeDst(27)
+	writeDst(28)
+	writeDst(29)
+	writeDst(30)
+	writeDst(31)
+	writeDst(32)
+	writeDst(33)
+	writeDst(34)
+	writeDst(35)
+	writeDst(36)
+	writeDst(37)
+	writeDst(38)
+	writeDst(39)
+	writeDst(40)
+	writeDst(41)
+	writeDst(42)
+	writeDst(43)
+	writeDst(44)
+	writeDst(45)
+	writeDst(46)
+	writeDst(47)
+	writeDst(48)
+	writeDst(49)
+	writeDst(50)
+	writeDst(51)
+	writeDst(52)
+	writeDst(53)
+	writeDst(54)
+	writeDst(55)
+	writeDst(56)
+	writeDst(57)
+	writeDst(58)
+	writeDst(59)
+	writeDst(60)
+	writeDst(61)
+	writeDst(62)
+	writeDst(63)
+
 	return nil
 }
 
diff -ruN /usr/lib/go/src/image/jpeg/writer_test.go ./optimized/jpeg/writer_test.go
--- /usr/lib/go/src/image/jpeg/writer_test.go	2025-02-05 01:18:32.000000000 +0100
+++ ./optimized/jpeg/writer_test.go	2025-02-07 20:16:25.281699218 +0100
@@ -20,7 +20,7 @@
 // zigzag maps from the natural ordering to the zig-zag ordering. For example,
 // zigzag[0*8 + 3] is the zig-zag sequence number of the element in the fourth
 // column and first row.
-var zigzag = [blockSize]int{
+var zigzag = [blockSize]uint8{
 	0, 1, 5, 6, 14, 15, 27, 28,
 	2, 4, 7, 13, 16, 26, 29, 42,
 	3, 8, 12, 17, 25, 30, 41, 43,
@@ -32,7 +32,7 @@
 }
 
 func TestZigUnzig(t *testing.T) {
-	for i := 0; i < blockSize; i++ {
+	for i := range uint8(blockSize) {
 		if unzig[zigzag[i]] != i {
 			t.Errorf("unzig[zigzag[%d]] == %d", i, unzig[zigzag[i]])
 		}
```

</details>

## Benchmark Results

### amd64 linux (Ryzen 7 5700X)

<details>

```
goos: linux
goarch: amd64
pkg: github.com/romshark/jpegbench
cpu: AMD Ryzen 7 5700X 8-Core Processor
                               │   std.txt   │               opt.txt               │
                               │   sec/op    │   sec/op     vs base                │
Decode/11375x8992_6mb.jpg-16     598.9m ± 0%   501.1m ± 0%  -16.32% (p=0.000 n=12)
Decode/1280x719_84kb.jpg-16      8.092m ± 0%   6.549m ± 0%  -19.07% (p=0.000 n=12)
Decode/15400x6940_20mb.jpg-16     1.950 ± 0%    1.626 ± 0%  -16.60% (p=0.000 n=12)
Decode/1920x1193_600kb.jpg-16    33.36m ± 0%   29.45m ± 0%  -11.74% (p=0.000 n=12)
Decode/32x32_.jpg-16             22.41µ ± 0%   21.31µ ± 0%   -4.92% (p=0.000 n=12)
Decode/6000x4000_2mb.jpg-16      358.0m ± 0%   316.3m ± 0%  -11.64% (p=0.000 n=12)
Decode/600x239_35kb.jpg-16       2.709m ± 0%   2.330m ± 0%  -13.98% (p=0.000 n=12)
Decode/9319x5792_6480kb.jpg-16   939.6m ± 0%   844.3m ± 1%  -10.15% (p=0.000 n=12)
geomean                          39.91m        34.66m       -13.15%

                               │   std.txt    │                opt.txt                │
                               │     B/op     │     B/op      vs base                 │
Decode/11375x8992_6mb.jpg-16     97.58Mi ± 0%   97.58Mi ± 0%       ~ (p=0.748 n=12)
Decode/1280x719_84kb.jpg-16      1.335Mi ± 0%   1.335Mi ± 0%       ~ (p=0.572 n=12)
Decode/15400x6940_20mb.jpg-16    306.0Mi ± 0%   306.0Mi ± 0%       ~ (p=1.000 n=12) ¹
Decode/1920x1193_600kb.jpg-16    3.310Mi ± 0%   3.310Mi ± 0%   0.00% (p=0.040 n=12)
Decode/32x32_.jpg-16             15.02Ki ± 0%   15.02Ki ± 0%       ~ (p=1.000 n=12) ¹
Decode/6000x4000_2mb.jpg-16      171.7Mi ± 0%   171.7Mi ± 0%       ~ (p=1.000 n=12)
Decode/600x239_35kb.jpg-16       437.5Ki ± 0%   437.5Ki ± 0%       ~ (p=1.000 n=12)
Decode/9319x5792_6480kb.jpg-16   386.5Mi ± 0%   386.5Mi ± 0%       ~ (p=1.000 n=12)
geomean                          9.276Mi        9.276Mi       +0.00%
¹ all samples are equal

                               │   std.txt   │               opt.txt                │
                               │  allocs/op  │  allocs/op   vs base                 │
Decode/11375x8992_6mb.jpg-16      626.0 ± 0%    626.0 ± 0%       ~ (p=1.000 n=12) ¹
Decode/1280x719_84kb.jpg-16       71.00 ± 0%    71.00 ± 0%       ~ (p=1.000 n=12) ¹
Decode/15400x6940_20mb.jpg-16    1.248k ± 0%   1.248k ± 0%       ~ (p=1.000 n=12) ¹
Decode/1920x1193_600kb.jpg-16     8.000 ± 0%    8.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/32x32_.jpg-16              5.000 ± 0%    5.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/6000x4000_2mb.jpg-16       11.00 ± 0%    11.00 ± 0%       ~ (p=1.000 n=12) ¹
Decode/600x239_35kb.jpg-16        6.000 ± 0%    6.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/9319x5792_6480kb.jpg-16    12.00 ± 0%    12.00 ± 0%       ~ (p=1.000 n=12)
geomean                           33.93         33.93       +0.00%
¹ all samples are equal

Binary sizes:
  optimized: 1967278 (1.9M)
  standard:  1951199 (1.9M)
diff:        16079 (16KiB)
```

</details>

### amd64 darwin (i7-4850HQ)

<details>

```
goos: darwin
goarch: amd64
pkg: github.com/romshark/jpegbench
cpu: Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz
                              │   std.txt    │               opt.txt               │
                              │    sec/op    │   sec/op     vs base                │
Decode/11375x8992_6mb.jpg-8      765.7m ± 1%   673.4m ± 0%  -12.05% (p=0.000 n=12)
Decode/1280x719_84kb.jpg-8      10.232m ± 2%   8.805m ± 1%  -13.95% (p=0.000 n=12)
Decode/15400x6940_20mb.jpg-8      2.456 ± 0%    2.138 ± 0%  -12.97% (p=0.000 n=12)
Decode/1920x1193_600kb.jpg-8     41.90m ± 1%   38.61m ± 1%   -7.87% (p=0.000 n=12)
Decode/32x32_.jpg-8              32.06µ ± 0%   30.96µ ± 2%   -3.45% (p=0.000 n=12)
Decode/6000x4000_2mb.jpg-8       481.9m ± 1%   442.9m ± 1%   -8.10% (p=0.000 n=12)
Decode/600x239_35kb.jpg-8        3.442m ± 0%   3.020m ± 0%  -12.24% (p=0.000 n=12)
Decode/9319x5792_6480kb.jpg-8     1.254 ± 1%    1.184 ± 1%   -5.60% (p=0.000 n=12)
geomean                          52.04m        47.04m        -9.60%

                              │   std.txt    │                opt.txt                │
                              │     B/op     │     B/op      vs base                 │
Decode/11375x8992_6mb.jpg-8     97.58Mi ± 0%   97.58Mi ± 0%       ~ (p=0.886 n=12)
Decode/1280x719_84kb.jpg-8      1.335Mi ± 0%   1.335Mi ± 0%  +0.00% (p=0.003 n=12)
Decode/15400x6940_20mb.jpg-8    306.0Mi ± 0%   306.0Mi ± 0%       ~ (p=1.000 n=12)
Decode/1920x1193_600kb.jpg-8    3.310Mi ± 0%   3.310Mi ± 0%  +0.00% (p=0.036 n=12)
Decode/32x32_.jpg-8             15.02Ki ± 0%   15.02Ki ± 0%       ~ (p=1.000 n=12) ¹
Decode/6000x4000_2mb.jpg-8      171.7Mi ± 0%   171.7Mi ± 0%       ~ (p=0.056 n=12)
Decode/600x239_35kb.jpg-8       437.5Ki ± 0%   437.5Ki ± 0%       ~ (p=0.119 n=12)
Decode/9319x5792_6480kb.jpg-8   386.5Mi ± 0%   386.5Mi ± 0%       ~ (p=0.479 n=12)
geomean                         9.276Mi        9.276Mi       +0.00%
¹ all samples are equal

                              │   std.txt   │               opt.txt                │
                              │  allocs/op  │  allocs/op   vs base                 │
Decode/11375x8992_6mb.jpg-8      626.0 ± 0%    626.0 ± 0%       ~ (p=1.000 n=12)
Decode/1280x719_84kb.jpg-8       71.00 ± 0%    71.00 ± 0%       ~ (p=1.000 n=12) ¹
Decode/15400x6940_20mb.jpg-8    1.248k ± 0%   1.248k ± 0%       ~ (p=1.000 n=12)
Decode/1920x1193_600kb.jpg-8     8.000 ± 0%    8.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/32x32_.jpg-8              5.000 ± 0%    5.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/6000x4000_2mb.jpg-8       11.00 ± 0%    11.00 ± 0%       ~ (p=1.000 n=12) ¹
Decode/600x239_35kb.jpg-8        6.000 ± 0%    6.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/9319x5792_6480kb.jpg-8    12.00 ± 8%    12.00 ± 0%       ~ (p=0.479 n=12)
geomean                          33.93         33.93       +0.00%
¹ all samples are equal

Binary sizes:
  optimized: 1975670 (1.9M)
  standard:  1959607 (1.9M)
diff:        16063 (16063B)
```

</details>

### arm64 darwin (M1)

<details>

```
goos: darwin
goarch: arm64
pkg: github.com/romshark/jpegbench
cpu: Apple M1 Max
                               │   std.txt   │               opt.txt               │
                               │   sec/op    │   sec/op     vs base                │
Decode/11375x8992_6mb.jpg-10     486.6m ± 0%   398.5m ± 1%  -18.11% (p=0.000 n=12)
Decode/1280x719_84kb.jpg-10      6.512m ± 0%   5.221m ± 0%  -19.82% (p=0.000 n=12)
Decode/15400x6940_20mb.jpg-10     1.566 ± 0%    1.277 ± 0%  -18.48% (p=0.000 n=12)
Decode/1920x1193_600kb.jpg-10    29.26m ± 0%   26.04m ± 0%  -11.01% (p=0.000 n=12)
Decode/32x32_.jpg-10             18.19µ ± 0%   16.69µ ± 0%   -8.27% (p=0.000 n=12)
Decode/6000x4000_2mb.jpg-10      279.8m ± 0%   252.1m ± 0%   -9.89% (p=0.000 n=12)
Decode/600x239_35kb.jpg-10       2.187m ± 0%   1.800m ± 0%  -17.71% (p=0.000 n=12)
Decode/9319x5792_6480kb.jpg-10   750.2m ± 0%   687.8m ± 0%   -8.32% (p=0.000 n=12)
geomean                          32.39m        27.83m       -14.08%

                               │   std.txt    │                opt.txt                │
                               │     B/op     │     B/op      vs base                 │
Decode/11375x8992_6mb.jpg-10     97.58Mi ± 0%   97.58Mi ± 0%       ~ (p=1.000 n=12)
Decode/1280x719_84kb.jpg-10      1.335Mi ± 0%   1.335Mi ± 0%       ~ (p=0.261 n=12)
Decode/15400x6940_20mb.jpg-10    306.0Mi ± 0%   306.0Mi ± 0%       ~ (p=1.000 n=12)
Decode/1920x1193_600kb.jpg-10    3.310Mi ± 0%   3.310Mi ± 0%  +0.00% (p=0.002 n=12)
Decode/32x32_.jpg-10             15.02Ki ± 0%   15.02Ki ± 0%       ~ (p=1.000 n=12) ¹
Decode/6000x4000_2mb.jpg-10      171.7Mi ± 0%   171.7Mi ± 0%       ~ (p=1.000 n=12)
Decode/600x239_35kb.jpg-10       437.5Ki ± 0%   437.5Ki ± 0%  +0.00% (p=0.026 n=12)
Decode/9319x5792_6480kb.jpg-10   386.5Mi ± 0%   386.5Mi ± 0%       ~ (p=1.000 n=12)
geomean                          9.276Mi        9.276Mi       +0.00%
¹ all samples are equal

                               │   std.txt   │               opt.txt                │
                               │  allocs/op  │  allocs/op   vs base                 │
Decode/11375x8992_6mb.jpg-10      626.0 ± 0%    626.0 ± 0%       ~ (p=1.000 n=12) ¹
Decode/1280x719_84kb.jpg-10       71.00 ± 0%    71.00 ± 0%       ~ (p=1.000 n=12) ¹
Decode/15400x6940_20mb.jpg-10    1.248k ± 0%   1.248k ± 0%       ~ (p=1.000 n=12)
Decode/1920x1193_600kb.jpg-10     8.000 ± 0%    8.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/32x32_.jpg-10              5.000 ± 0%    5.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/6000x4000_2mb.jpg-10       11.00 ± 0%    11.00 ± 0%       ~ (p=1.000 n=12) ¹
Decode/600x239_35kb.jpg-10        6.000 ± 0%    6.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/9319x5792_6480kb.jpg-10    12.00 ± 0%    12.00 ± 0%       ~ (p=1.000 n=12)
geomean                           33.93         33.93       +0.00%
¹ all samples are equal

Binary sizes:
  optimized: 1975694 (1.9M)
  standard:  1959599 (1.9M)
diff:        16095 (16095B)
```

</details>

### arm64 darwin (M4)

<details>

```
goos: darwin
goarch: arm64
pkg: github.com/romshark/jpegbench
cpu: Apple M4 Pro
                               │   std.txt    │               opt.txt               │
                               │    sec/op    │   sec/op     vs base                │
Decode/11375x8992_6mb.jpg-14      321.0m ± 1%   277.6m ± 0%  -13.52% (p=0.000 n=12)
Decode/1280x719_84kb.jpg-14       4.169m ± 0%   3.394m ± 0%  -18.60% (p=0.000 n=12)
Decode/15400x6940_20mb.jpg-14    1064.5m ± 0%   900.9m ± 0%  -15.37% (p=0.000 n=12)
Decode/1920x1193_600kb.jpg-14     20.94m ± 0%   19.44m ± 0%   -7.19% (p=0.000 n=12)
Decode/32x32_.jpg-14              13.91µ ± 0%   13.16µ ± 0%   -5.40% (p=0.000 n=12)
Decode/6000x4000_2mb.jpg-14       183.7m ± 1%   169.4m ± 1%   -7.81% (p=0.000 n=12)
Decode/600x239_35kb.jpg-14        1.417m ± 0%   1.158m ± 1%  -18.30% (p=0.000 n=12)
Decode/9319x5792_6480kb.jpg-14    504.3m ± 1%   475.6m ± 0%   -5.70% (p=0.000 n=12)
geomean                           21.98m        19.42m       -11.64%

                               │   std.txt    │                opt.txt                │
                               │     B/op     │     B/op      vs base                 │
Decode/11375x8992_6mb.jpg-14     97.58Mi ± 0%   97.58Mi ± 0%       ~ (p=0.120 n=12)
Decode/1280x719_84kb.jpg-14      1.335Mi ± 0%   1.335Mi ± 0%  -0.00% (p=0.006 n=12)
Decode/15400x6940_20mb.jpg-14    306.0Mi ± 0%   306.0Mi ± 0%       ~ (p=1.000 n=12)
Decode/1920x1193_600kb.jpg-14    3.310Mi ± 0%   3.310Mi ± 0%       ~ (p=0.919 n=12)
Decode/32x32_.jpg-14             15.02Ki ± 0%   15.02Ki ± 0%       ~ (p=1.000 n=12) ¹
Decode/6000x4000_2mb.jpg-14      171.7Mi ± 0%   171.7Mi ± 0%       ~ (p=0.733 n=12)
Decode/600x239_35kb.jpg-14       437.5Ki ± 0%   437.5Ki ± 0%       ~ (p=0.217 n=12)
Decode/9319x5792_6480kb.jpg-14   386.5Mi ± 0%   386.5Mi ± 0%       ~ (p=0.590 n=12)
geomean                          9.276Mi        9.276Mi       -0.00%
¹ all samples are equal

                               │   std.txt   │               opt.txt                │
                               │  allocs/op  │  allocs/op   vs base                 │
Decode/11375x8992_6mb.jpg-14      626.0 ± 0%    626.0 ± 0%       ~ (p=1.000 n=12) ¹
Decode/1280x719_84kb.jpg-14       71.00 ± 0%    71.00 ± 0%       ~ (p=1.000 n=12) ¹
Decode/15400x6940_20mb.jpg-14    1.248k ± 0%   1.248k ± 0%       ~ (p=1.000 n=12) ¹
Decode/1920x1193_600kb.jpg-14     8.000 ± 0%    8.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/32x32_.jpg-14              5.000 ± 0%    5.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/6000x4000_2mb.jpg-14       11.00 ± 0%    11.00 ± 0%       ~ (p=1.000 n=12) ¹
Decode/600x239_35kb.jpg-14        6.000 ± 0%    6.000 ± 0%       ~ (p=1.000 n=12) ¹
Decode/9319x5792_6480kb.jpg-14    12.00 ± 0%    12.00 ± 0%       ~ (p=1.000 n=12) ¹
geomean                           33.93         33.93       +0.00%
¹ all samples are equal

Binary sizes:
  optimized: 1975662 (1.9M)
  standard:  1959591 (1.9M)
diff:        16071 (16071B)
```

</details>

## Running tests

```sh
go test -v ./...
```

## Running benchmark

```sh
./bench.sh 12 . && ./cmpbinsz.sh
```
