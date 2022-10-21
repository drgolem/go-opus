# go-opus

Go bindings for libopus


WIP, learning cgo and audio libs

Example:
```sh
go run examples/ogg2raw.go 1.ogg out_1.raw
```
play raw samples:
```sh
ffplay -ar 48000 -ac 2 -f s16le out_1.raw
```


Check correctness:
```sh
ffmpeg -i 1.ogg -f s16le -acodec pcm_s16le ref_1.raw
go run examples/ogg2raw.go 1.ogg out_1.raw
diff ref_1.raw out_1.raw
```
