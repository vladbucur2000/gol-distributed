package gol

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events      chan<- Event
	ioCommand   chan<- ioCommand
	ioIdle      <-chan bool
	ioFilename  chan<- string
	ioInput     <-chan uint8
	ioOutput    chan<- uint8
	outputWorld chan<- [][]byte
}

func read(c distributorChannels, conn *net.Conn) {
	reader := bufio.NewReader(*conn)
	for {
		msg, _ := reader.ReadString('\n')
		//CELLFLIPPED RECEIVED
		if msg[0] == 'c' && msg[1] == 'f' {
			i := 2
			x := 0
			y := 0
			turn := 0
			for i < len(msg) && msg[i] != ' ' {
				y = y*10 + (int(msg[i]) - '0')
				i++
			}
			i++
			for i < len(msg) && msg[i] != ' ' {
				x = x*10 + (int(msg[i]) - '0')
				i++
			}
			i++
			for i < len(msg) && msg[i] != '\n' {
				turn = turn*10 + (int(msg[i]) - '0')
				i++
			}
			c.events <- CellFlipped{
				turn,
				util.Cell{
					X: x,
					Y: y,
				},
			}
			//	break
			//TURN COMPLETE
		} else if msg[0] == 't' && msg[1] == 'c' {
			i := 2
			turn := 0
			for i < len(msg) && msg[i] != '\n' {
				turn = turn*10 + (int(msg[i]) - '0')
				i++
			}
			c.events <- TurnComplete{
				turn,
			}

			//FINAL TURN COMPLETE
		} else if msg[0] == 'f' && msg[1] == 't' && msg[2] == 'c' {

			//AICI TREBUIE MODIFICAT PENTRU "BUG"
			i := 3
			turn := 0
			var alive []util.Cell
			for i < len(msg) && msg[i] != ' ' {
				turn = turn*10 + (int(msg[i]) - '0')
				i++
			}

			i += 3
			for i < len(msg) && msg[i] != '\n' {
				x := 0
				y := 0
				j := i

				for j < len(msg) && msg[j] != ';' {
					x = x*10 + (int(msg[j]) - '0')
					j++
				}

				j += 3

				for j < len(msg) && msg[j] != ' ' {
					y = y*10 + (int(msg[j]) - '0')
					j++
				}

				alive = append(alive, util.Cell{
					X: x,
					Y: y,
				})
				j += 3
				i = j
			}

			c.events <- FinalTurnComplete{
				turn,
				alive,
			}
			return

		}
	}

}

func numberToString(nr int) string {
	return strconv.Itoa(nr)
}

func convertToString(world [][]byte, p Params) string {
	var data []string

	hs := numberToString(p.ImageHeight)
	ws := numberToString(p.ImageWidth)
	turn := numberToString(p.Turns)
	thread := numberToString(p.Threads)

	data = append(data, hs)
	data = append(data, "\n")
	data = append(data, ws)
	data = append(data, "\n")
	data = append(data, turn)
	data = append(data, "\n")
	data = append(data, thread)
	data = append(data, "\n")

	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {

			if world[i][j] == 255 {
				data = append(data, "1")
			}
			if world[i][j] == 0 {
				data = append(data, "0")
			}

		}
		data = append(data, "\n")
	}

	data = append(data, "gata!!\t")

	return strings.Join(data, "")
}

// distributor divides the work between workers and interacts with other goroutines.
func controller(p Params, c distributorChannels) {

	conn, _ := net.Dial("tcp", "127.0.0.1:8080")

	//citesc lumea
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	c.ioCommand <- ioInput

	c.ioFilename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")
	// TODO: For all initially alive cells send a CellFlipped Event.
	// SA EXTRAG INPUTUL IN MATRICE SI SA L TRANSFORM IN STRING CA SA L TRIMIT LA ENGINE
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			val := <-c.ioInput
			world[i][j] = val
		}
	}

	//text := util.MatricesToString(world, nil, p.ImageWidth, p.ImageHeight)
	text := convertToString(world, p)

	fmt.Fprintf(conn, text)
	//for {
	read(c, &conn)
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	//	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
