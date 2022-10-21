package main

import (
	"fmt"
	"os"

	"github.com/drgolem/go-opus/opus"
)

func main() {
	fmt.Println("example decode opus to wav file")

	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: ogg2raw <infile.ogg> <outfile.raw>")
		fmt.Fprintln(os.Stderr, "play: ffplay -ar 48000 -ac 2 -f s16le <outfile.raw>")
		return
	}

	inFile := os.Args[1]
	outFile := os.Args[2]
	fmt.Printf("infile: %s, outfile: %s\n", inFile, outFile)

	outBitsPerSample := 16
	outBytesPerSample := outBitsPerSample / 8

	dec, err := opus.NewOpusFileDecoder(inFile)
	if err != nil {
		panic(err)
	}
	defer dec.Close()

	fmt.Printf("current sample: %d\n", dec.TellCurrentSample())
	fmt.Printf("total samples: %d\n", dec.TotalSamples())

	rate, channels, bitsPerSample := dec.GetFormat()
	fmt.Printf("Format: [%d:%d:%d]\n", rate, channels, bitsPerSample)

	fOut, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
	}
	defer fOut.Close()

	audioSamples := 5760 * channels
	audioBufferBytes := audioSamples * 2
	audio := make([]byte, audioBufferBytes)

	for {
		sampleCnt, err := dec.DecodeSamples(audioSamples, audio)
		if err != nil {
			fmt.Printf("ERR: %v\n", err)
			break
		}
		if sampleCnt == 0 {
			break
		}

		bytesToWrite := sampleCnt * channels * outBytesPerSample
		fOut.Write(audio[:bytesToWrite])
	}
	fOut.Sync()
	fmt.Printf("current sample: %d\n", dec.TellCurrentSample())
}
