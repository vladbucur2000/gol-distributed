package gol

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"uk.ac.bris.cs/gameoflife/util"
)

type myParameters struct {
	clientid    int
	world       [][]byte
	ImageHeight int
	ImageWidth  int
	Turns       int
	Threads     int
}

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioInput    <-chan uint8
	ioOutput   chan<- uint8
	keyPressed <-chan rune
}

func sendKeys(c distributorChannels, conn *net.Conn, keyTurn chan string) {
	for {
		select {
		case key := <-c.keyPressed:
			if key == 'p' {
				text := "kpauseTheGame\n"
				fmt.Fprintf(*conn, text)

			} else if key == 's' {
				fmt.Println("Saving key sent to server...")
				text := "ksaveTheGame\n"
				fmt.Fprintf(*conn, text)

			} else if key == 'q' {
				fmt.Println("Quiting key sent to server...")
				text := "kquitTheGame\n"
				fmt.Fprintf(*conn, text)
			}

		default:
		}
	}
}
func saveTheWorld(c distributorChannels, p myParameters) {
	c.ioCommand <- ioOutput
	c.ioFilename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight), strconv.Itoa(p.Turns)}, "x")
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			c.ioOutput <- p.world[i][j]
		}
	}
}
func read(c distributorChannels, conn *net.Conn, keyTurn chan string) {
	reader := bufio.NewReader(*conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		if len(msg) == 1 {
			continue
		}

		fmt.Println(msg)

		if msg[1] == 'm' && msg[2] == 'a' && msg[3] == 'p' {

			p := stringToMatrix(msg)
			saveTheWorld(c, p)

		} else if msg[0] == 'k' && msg[1] == 'e' && msg[2] == 'y' {

			var state State = 0
			if msg[3] == 'p' {
				state = Paused
			} else if msg[3] == 'e' {
				state = 1
			} else if msg[3] == 'q' {
				state = 2
			}
			i := 4
			turn := 0
			for i < len(msg) && msg[i] != ' ' {
				turn = turn*10 + (int(msg[i]) - '0')
				i++
			}
			c.events <- StateChange{
				turn,
				state,
			}
			if state == 2 {
				//return
			}
			//CELLFLIPPED RECEIVED
		} else if msg[0] == 'a' && msg[1] == 'c' && msg[2] == 'c' {
			i := 3
			turn := 0
			howManyAreComplete := 0
			for i < len(msg) && msg[i] != ' ' {
				turn = turn*10 + (int(msg[i]) - '0')
				i++
			}
			i++
			for i < len(msg) && msg[i] != '\n' {
				howManyAreComplete = howManyAreComplete*10 + (int(msg[i]) - '0')
				i++
			}
			//fmt.Println(howManyAreComplete)
			c.events <- AliveCellsCount{
				turn,
				howManyAreComplete,
			}
		} else if msg[0] == 'c' && msg[1] == 'f' {
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
	data = append(data, "0map")
	data = append(data, hs)
	data = append(data, " ")
	data = append(data, ws)
	data = append(data, " ")
	data = append(data, turn)
	data = append(data, " ")
	data = append(data, thread)
	data = append(data, " ")

	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {

			if world[i][j] == 255 {
				data = append(data, "1")
			}
			if world[i][j] == 0 {
				data = append(data, "0")
			}

		}
		data = append(data, " ")
	}

	data = append(data, "\n")

	return strings.Join(data, "")
}
func stringToMatrix(msg string) myParameters {
	clientid := (int(msg[0]) - '0')
	i := 4
	height := 0
	width := 0
	turn := 0
	thread := 0

	for i < len(msg) && msg[i] != ' ' {
		height = height*10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for i < len(msg) && msg[i] != ' ' {
		width = width*10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for i < len(msg) && msg[i] != ' ' {
		turn = turn*10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for i < len(msg) && msg[i] != ' ' {
		thread = thread*10 + (int(msg[i]) - '0')
		i++
	}

	i++
	nr := i

	world := make([][]byte, height)
	for i := range world {
		world[i] = make([]byte, width)
	}

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if msg[nr] == '0' {
				world[i][j] = 0
			} else {
				world[i][j] = 255
			}

			nr++

		}
		nr++
	}

	return myParameters{
		clientid,
		world,
		height,
		width,
		turn,
		thread,
	}
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
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			val := <-c.ioInput
			world[i][j] = val
		}
	}

	text := convertToString(world, p)

	fmt.Fprintf(conn, text)
	keyTurn := make(chan string)

	go sendKeys(c, &conn, keyTurn)
	read(c, &conn, keyTurn)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	//	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
