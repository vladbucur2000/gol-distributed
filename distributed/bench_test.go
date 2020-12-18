package main

import (
	"fmt"
	"os"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

func BenchmarkGol(b *testing.B) {
	fmt.Println("salut")

	os.Stdout = nil // Disable all program output apart from benchmark results
	tests := []gol.Params{
		//{ImageWidth: 16, ImageHeight: 16},
		{ImageWidth: 64, ImageHeight: 64},
		//{ImageWidth: 5120, ImageHeight: 5120},
	}
	for _, p := range tests {
		for _, turns := range []int{100} {
			p.Turns = turns

			for threads := 1; threads <= 16; threads++ {
				p.Threads = threads
				testName := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)
				b.Run(testName, func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						events := make(chan gol.Event)
						b.StartTimer()
						gol.Run(p, events, nil)

						for event := range events {
							switch event.(type) {
							case gol.FinalTurnComplete:
								b.StopTimer()
							}
						}
					}

				})
			}
		}
	}

}
