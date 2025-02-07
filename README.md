# Go's standard `image/jpeg` decoding optimization.

This repository demonstrates a potential optimization of decoding JPEG images
in the Go standard library. This is achieved by unrolling
unzig and dst writing loops in `scan.go`.

On average across benchmarks (Go 1.23.6), this is **~13% faster on AMD64**
and **~14% faster on ARM64 (M1)** at the cost of slightly larger binary size (~18KiB).

**amd64 linux**

<details>

```
goos: linux
goarch: amd64
pkg: github.com/romshark/jpegbench
cpu: AMD Ryzen 7 5700X 8-Core Processor             
                               │   std.txt   │               opt.txt               │
                               │   sec/op    │   sec/op     vs base                │
Decode/11375x8992_6mb.jpg-16     600.6m ± 0%   505.0m ± 1%  -15.91% (p=0.000 n=10)
Decode/1280x719_84kb.jpg-16      8.090m ± 0%   6.596m ± 0%  -18.47% (p=0.000 n=10)
Decode/15400x6940_20mb.jpg-16     1.959 ± 0%    1.640 ± 0%  -16.27% (p=0.000 n=10)
Decode/1920x1193_600kb.jpg-16    33.38m ± 0%   29.84m ± 0%  -10.59% (p=0.000 n=10)
Decode/32x32_.jpg-16             22.41µ ± 0%   20.60µ ± 0%   -8.06% (p=0.000 n=10)
Decode/6000x4000_2mb.jpg-16      359.2m ± 0%   317.8m ± 0%  -11.53% (p=0.000 n=10)
Decode/600x239_35kb.jpg-16       2.707m ± 0%   2.427m ± 0%  -10.34% (p=0.000 n=10)
Decode/9319x5792_6480kb.jpg-16   939.3m ± 0%   845.8m ± 0%   -9.95% (p=0.000 n=10)
geomean                          39.96m        34.88m       -12.71%

                               │   std.txt    │                opt.txt                │
                               │     B/op     │     B/op      vs base                 │
Decode/11375x8992_6mb.jpg-16     97.58Mi ± 0%   97.58Mi ± 0%       ~ (p=0.433 n=10)
Decode/1280x719_84kb.jpg-16      1.335Mi ± 0%   1.335Mi ± 0%       ~ (p=0.897 n=10)
Decode/15400x6940_20mb.jpg-16    306.0Mi ± 0%   306.0Mi ± 0%       ~ (p=1.000 n=10)
Decode/1920x1193_600kb.jpg-16    3.310Mi ± 0%   3.310Mi ± 0%  +0.00% (p=0.004 n=10)
Decode/32x32_.jpg-16             15.02Ki ± 0%   15.02Ki ± 0%       ~ (p=1.000 n=10) ¹
Decode/6000x4000_2mb.jpg-16      171.7Mi ± 0%   171.7Mi ± 0%       ~ (p=0.373 n=10)
Decode/600x239_35kb.jpg-16       437.5Ki ± 0%   437.5Ki ± 0%       ~ (p=1.000 n=10) ¹
Decode/9319x5792_6480kb.jpg-16   386.5Mi ± 0%   386.5Mi ± 0%       ~ (p=0.474 n=10)
geomean                          9.276Mi        9.276Mi       +0.00%
¹ all samples are equal

                               │   std.txt   │               opt.txt                │
                               │  allocs/op  │  allocs/op   vs base                 │
Decode/11375x8992_6mb.jpg-16      626.0 ± 0%    626.0 ± 0%       ~ (p=1.000 n=10)
Decode/1280x719_84kb.jpg-16       71.00 ± 0%    71.00 ± 0%       ~ (p=1.000 n=10) ¹
Decode/15400x6940_20mb.jpg-16    1.248k ± 0%   1.248k ± 0%       ~ (p=1.000 n=10)
Decode/1920x1193_600kb.jpg-16     8.000 ± 0%    8.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/32x32_.jpg-16              5.000 ± 0%    5.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/6000x4000_2mb.jpg-16       11.00 ± 0%    11.00 ± 0%       ~ (p=1.000 n=10) ¹
Decode/600x239_35kb.jpg-16        6.000 ± 0%    6.000 ± 0%       ~ (p=1.000 n=10) ¹
Decode/9319x5792_6480kb.jpg-16    12.00 ± 0%    12.00 ± 0%       ~ (p=1.000 n=10)
geomean                           33.93         33.93       +0.00%
¹ all samples are equal

Binary sizes:
  optimized: 1968985 (1.9M)
  standard:  1951199 (1.9M)
diff:        17786 (18KiB)
```

</details>

**arm64 darwin**

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

## Running tests

```sh
go test -v ./...
```

## Running benchmark

```sh
./bench.sh 10 .
```

## Running binary size comparison

```sh
./cmpbinsz.sh
```
