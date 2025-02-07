# Go's standard `image/jpeg` decoding optimization.

This repository demonstrates a potential optimization of decoding JPEG images
in the Go standard library. This is achieved by unrolling
unzig and shift-clamp loops in `scan.go`.

On average across benchmarks (Go 1.23.6), this is **~13% faster on AMD64**
and **~14%/11% faster on Apple Silicon ARM64 (M1/M4)**
at the cost of slightly larger binary size (~16KiB).

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
Decode/11375x8992_6mb.jpg-8      766.2m ± 1%   682.4m ± 0%  -10.93% (p=0.000 n=10)
Decode/1280x719_84kb.jpg-8      10.197m ± 0%   8.938m ± 0%  -12.35% (p=0.000 n=10)
Decode/15400x6940_20mb.jpg-8      2.463 ± 2%    2.177 ± 0%  -11.58% (p=0.000 n=10)
Decode/1920x1193_600kb.jpg-8     42.01m ± 1%   38.99m ± 0%   -7.18% (p=0.000 n=10)
Decode/32x32_.jpg-8              31.93µ ± 0%   30.59µ ± 0%   -4.20% (p=0.000 n=10)
Decode/6000x4000_2mb.jpg-8       480.6m ± 1%   441.8m ± 1%   -8.07% (p=0.000 n=10)
Decode/600x239_35kb.jpg-8        3.446m ± 0%   3.057m ± 1%  -11.29% (p=0.000 n=10)
Decode/9319x5792_6480kb.jpg-8     1.267 ± 1%    1.199 ± 6%   -5.36% (p=0.019 n=10)
geomean                          52.08m        47.44m        -8.92%

                              │   std.txt    │                opt.txt                │
                              │     B/op     │     B/op      vs base                 │
Decode/11375x8992_6mb.jpg-8     97.58Mi ± 0%   97.58Mi ± 0%       ~ (p=0.704 n=10)
Decode/1280x719_84kb.jpg-8      1.335Mi ± 0%   1.335Mi ± 0%       ~ (p=0.467 n=10)
Decode/15400x6940_20mb.jpg-8    306.0Mi ± 0%   306.0Mi ± 0%       ~ (p=0.721 n=10)
Decode/1920x1193_600kb.jpg-8    3.310Mi ± 0%   3.310Mi ± 0%       ~ (p=0.515 n=10)
Decode/32x32_.jpg-8             15.02Ki ± 0%   15.02Ki ± 0%       ~ (p=1.000 n=10) ¹
Decode/6000x4000_2mb.jpg-8      171.7Mi ± 0%   171.7Mi ± 0%       ~ (p=1.000 n=10)
Decode/600x239_35kb.jpg-8       437.5Ki ± 0%   437.5Ki ± 0%       ~ (p=0.657 n=10)
Decode/9319x5792_6480kb.jpg-8   386.5Mi ± 0%   386.5Mi ± 0%       ~ (p=0.308 n=10)
geomean                         9.276Mi        9.276Mi       -0.00%
¹ all samples are equal

                              │   std.txt    │               opt.txt                │
                              │  allocs/op   │  allocs/op   vs base                 │
Decode/11375x8992_6mb.jpg-8      626.0 ±  0%    626.0 ± 0%       ~ (p=0.474 n=10)
Decode/1280x719_84kb.jpg-8       71.00 ±  0%    71.00 ± 0%       ~ (p=1.000 n=10) ¹
Decode/15400x6940_20mb.jpg-8    1.248k ±  0%   1.248k ± 0%       ~ (p=0.721 n=10)
Decode/1920x1193_600kb.jpg-8     8.000 ±  0%    8.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/32x32_.jpg-8              5.000 ±  0%    5.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/6000x4000_2mb.jpg-8       11.00 ±  0%    11.00 ± 0%       ~ (p=1.000 n=10) ¹
Decode/600x239_35kb.jpg-8        6.000 ±  0%    6.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/9319x5792_6480kb.jpg-8    12.50 ± 12%    12.00 ± 8%       ~ (p=0.308 n=10)
geomean                          34.11          33.93       -0.51%
¹ all samples are equal

Binary sizes:
  optimized: 1977345 (1.9M)
  standard:  1959607 (1.9M)
diff:        17738 (17738B)
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
Decode/11375x8992_6mb.jpg-10     486.8m ± 1%   399.3m ± 1%  -17.98% (p=0.000 n=10)
Decode/1280x719_84kb.jpg-10      6.510m ± 0%   5.224m ± 0%  -19.76% (p=0.000 n=10)
Decode/15400x6940_20mb.jpg-10     1.567 ± 0%    1.280 ± 1%  -18.36% (p=0.000 n=10)
Decode/1920x1193_600kb.jpg-10    29.28m ± 0%   26.04m ± 0%  -11.08% (p=0.000 n=10)
Decode/32x32_.jpg-10             18.23µ ± 0%   16.70µ ± 0%   -8.40% (p=0.000 n=10)
Decode/6000x4000_2mb.jpg-10      279.9m ± 0%   252.0m ± 0%   -9.98% (p=0.000 n=10)
Decode/600x239_35kb.jpg-10       2.183m ± 0%   1.804m ± 0%  -17.39% (p=0.000 n=10)
Decode/9319x5792_6480kb.jpg-10   748.3m ± 0%   689.8m ± 0%   -7.81% (p=0.000 n=10)
geomean                          32.39m        27.87m       -13.97%

                               │   std.txt    │                opt.txt                │
                               │     B/op     │     B/op      vs base                 │
Decode/11375x8992_6mb.jpg-10     97.58Mi ± 0%   97.58Mi ± 0%       ~ (p=0.799 n=10)
Decode/1280x719_84kb.jpg-10      1.335Mi ± 0%   1.335Mi ± 0%       ~ (p=0.123 n=10)
Decode/15400x6940_20mb.jpg-10    306.0Mi ± 0%   306.0Mi ± 0%       ~ (p=1.000 n=10)
Decode/1920x1193_600kb.jpg-10    3.310Mi ± 0%   3.310Mi ± 0%       ~ (p=0.365 n=10)
Decode/32x32_.jpg-10             15.02Ki ± 0%   15.02Ki ± 0%       ~ (p=1.000 n=10) ¹
Decode/6000x4000_2mb.jpg-10      171.7Mi ± 0%   171.7Mi ± 0%       ~ (p=0.303 n=10)
Decode/600x239_35kb.jpg-10       437.5Ki ± 0%   437.5Ki ± 0%       ~ (p=0.122 n=10)
Decode/9319x5792_6480kb.jpg-10   386.5Mi ± 0%   386.5Mi ± 0%       ~ (p=0.474 n=10)
geomean                          9.276Mi        9.276Mi       -0.00%
¹ all samples are equal

                               │   std.txt   │               opt.txt                │
                               │  allocs/op  │  allocs/op   vs base                 │
Decode/11375x8992_6mb.jpg-10      626.0 ± 0%    626.0 ± 0%       ~ (p=1.000 n=10)
Decode/1280x719_84kb.jpg-10       71.00 ± 0%    71.00 ± 0%       ~ (p=1.000 n=10) ¹
Decode/15400x6940_20mb.jpg-10    1.248k ± 0%   1.248k ± 0%       ~ (p=1.000 n=10)
Decode/1920x1193_600kb.jpg-10     8.000 ± 0%    8.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/32x32_.jpg-10              5.000 ± 0%    5.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/6000x4000_2mb.jpg-10       11.00 ± 0%    11.00 ± 0%       ~ (p=1.000 n=10) ¹
Decode/600x239_35kb.jpg-10        6.000 ± 0%    6.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/9319x5792_6480kb.jpg-10    12.00 ± 0%    12.00 ± 0%       ~ (p=1.000 n=10) ¹
geomean                           33.93         33.93       +0.00%
¹ all samples are equal

Binary sizes:
  optimized: 1973185 (1.9M)
  standard:  1955439 (1.9M)
diff:        17746 (17746B)
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
