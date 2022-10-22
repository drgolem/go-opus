package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/drgolem/go-ogg/ogg"
	"github.com/drgolem/go-opus/opus"
)

var opusHeadPattern = [8]byte{'O', 'p', 'u', 's', 'H', 'e', 'a', 'd'}
var opusTagsPattern = [8]byte{'O', 'p', 'u', 's', 'T', 'a', 'g', 's'}

type StreamType int

const (
	StreamType_Unknown StreamType = iota
	StreamType_Opus
)

type opusCommonHeader struct {
	OpusPattern [8]byte
}

type opusIdentificationHeader struct {
	Version         byte
	AudioChannels   byte
	PreSkip         uint16
	AudioSampleRate uint32
	OutputGain      uint16
	MappingFamily   byte
}

func main() {
	fmt.Println("example decode opus to wav file")

	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: opus2raw <infile.ogg> <outfile.raw>")
		fmt.Fprintln(os.Stderr, "play: ffplay -ar 48000 -ac 2 -f s16le <outfile.raw>")
		return
	}

	inFile := os.Args[1]
	outFile := os.Args[2]

	fmt.Printf("infile: %s, outfile: %s\n", inFile, outFile)

	f, err := os.Open(inFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	oggReader, err := ogg.NewOggReader(reader)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
		return
	}

	pktCnt := 0
	var streamType StreamType
	streamHeaders := make([][]byte, 0)
	headersCount := 0
	for oggReader.Next() {
		p, err := oggReader.Scan()
		if err != nil {
			fmt.Printf("ERR: %v\n", err)
			return
		}
		//fmt.Printf("packet (len: %d) data:\n%s\n", len(p), hex.Dump(p))
		pktCnt++

		if streamType == StreamType_Unknown {
			bytesReader := bytes.NewReader(p)
			var coh opusCommonHeader
			err = binary.Read(bytesReader, binary.LittleEndian, &coh)
			if err == nil {
				if coh.OpusPattern == opusHeadPattern {
					streamType = StreamType_Opus
				}
			}
			fmt.Printf("stream type: %v\n", streamType)

			switch streamType {
			case StreamType_Opus:
				headersCount = 2
			}
		}

		switch streamType {
		case StreamType_Opus:
			streamHeaders = append(streamHeaders, p)
			headersCount--
		}

		if headersCount == 0 {
			break
		}
	}

	outBitsPerSample := 16
	outBytesPerSample := outBitsPerSample / 8

	channels := 2
	sampleRate := 48000

	dec, err := opus.NewOpusPacketDecoder(channels, sampleRate)
	if err != nil {
		panic(err)
	}
	defer dec.Close()

	decOut, err := opus.NewOpusPacketDecoder(channels, sampleRate)
	if err != nil {
		panic(err)
	}
	defer decOut.Close()

	rate, channels, bitsPerSample := dec.GetFormat()
	fmt.Printf("Format: [%d:%d:%d]\n", rate, channels, bitsPerSample)

	fOut, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
	}
	defer fOut.Close()

	audioSamples := 5760 * channels
	audioBufferBytes := audioSamples * outBytesPerSample
	audio := make([]byte, audioBufferBytes)

	audioOut := make([]byte, audioBufferBytes)

	enc, err := opus.NewOpusPacketEncoder(channels)
	if err != nil {
		panic(err)
	}
	defer enc.Close()

	totalSamples := 0
	totalSamplesOut := 0
	for oggReader.Next() {
		packet, err := oggReader.Scan()
		if err != nil {
			fmt.Printf("ERR: %v\n", err)
			return
		}
		//fmt.Printf("packet (len: %d) data:\n%s\n", len(p), hex.Dump(p))
		pktCnt++

		sampleCnt, err := dec.DecodeSamples(packet, audioSamples, audio)
		if err != nil {
			fmt.Printf("ERR: %v\n", err)
			break
		}
		if sampleCnt == 0 {
			break
		}
		totalSamples += sampleCnt

		bytesToWrite := sampleCnt * channels * outBytesPerSample

		packetOut := make([]byte, audioBufferBytes)
		encodedBytes, err := enc.EncodeSamples(packetOut, sampleCnt, audio[:bytesToWrite])
		if err != nil {
			panic(err)
		}

		sampleCntOut, err := decOut.DecodeSamples(packetOut[:encodedBytes], sampleCnt, audioOut)
		if err != nil {
			fmt.Printf("ERR: %v\n", err)
			break
		}
		if sampleCntOut == 0 {
			break
		}
		totalSamplesOut += sampleCntOut

		bytesToWriteOut := sampleCntOut * channels * outBytesPerSample
		fOut.Write(audioOut[:bytesToWriteOut])
	}
	fOut.Sync()

	fmt.Printf("total decoded samples: %d\n", totalSamples)
	fmt.Printf("total decoded samples (dec2enc): %d\n", totalSamplesOut)
}
