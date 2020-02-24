package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/jfreymuth/pulse"
)

var wavFile *wav.Decoder
var done chan bool
var ticker *time.Ticker
var timePassed time.Duration

func printTime() {
	d, _ := wavFile.Duration()
	to := time.Now().Add(d)
	for {
		select {
		case <-ticker.C:
			// TODO: calculate the time better, it just miscalculate it
			timePassed = time.Since(to)
			fmt.Printf("\r%02d:%02d/%02d:%02d",
				int64(math.Abs(timePassed.Minutes())),
				int64(math.Abs(timePassed.Seconds())),
				int64(d.Minutes()), int64(d.Seconds()),
			)
		}
	}
}

func getWavBuf(out []int32) {

	if wavFile.EOF() {
		fmt.Println("EOF")
		done <- true
		return
	}

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

	l := len(buf.Data)
	if n != l {
		buf.Data = buf.Data[:n]
	}

	for i := range out {
		if i < l {
			out[i] = int32(buf.Data[i]) * int32(wavFile.SampleRate)
		}
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
	ticker = time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()
	go printTime()

	fmt.Printf("playing... %s\n", wavFile)
	<-done
	fmt.Println("done")

	stream.Close()
	close(done)
}
