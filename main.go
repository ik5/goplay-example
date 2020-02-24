package main

import (
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/jfreymuth/pulse"
)

var wavFile *wav.Decoder
var done chan bool

func getWavBuf(out []int32) {

	bufferSize := len(out)
	buf := &audio.IntBuffer{Data: make([]int, bufferSize), Format: wavFile.Format()}
	n, err := wavFile.PCMBuffer(buf)

	if err != nil {
		panic(err)
	}

	if n == 0 {
		fmt.Println("n 0")
		done <- true
		return
	}

	if n != len(buf.Data) {
		buf.Data = buf.Data[:n]
	}

	for i := range out {
		out[i] = int32(buf.Data[i]) * int32(wavFile.SampleRate)
	}
}

func main() {
	file, err := os.Open("wave/admiralbob77_-_Into_the_J_1.wav")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	wavFile = wav.NewDecoder(file)
	if !wavFile.IsValidFile() {
		fmt.Println("invalid WAV file")
		os.Exit(1)
	}

	c, err := pulse.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	s, err := c.DefaultSink()
	if err != nil {
		fmt.Println(err)
		return
	}

	wavFile.ReadInfo()
	done = make(chan bool)
	stream, err := c.NewPlayback(getWavBuf, pulse.PlaybackSink(s))
	if err != nil {
		fmt.Println(err)
		return
	}

	stream.Start()
	fmt.Printf("playing... %s\n", wavFile)
	<-done
	fmt.Println("done")

	stream.Close()
	close(done)
}
