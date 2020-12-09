package gol

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

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
				fmt.Println("Pause key sent to server...")
				text := "keypauseTheGame\n"
				fmt.Fprintf(*conn, text)
			} else if key == 's' {
				fmt.Println("Save key sent to server...")
				text := "keysaveTheGame\n"
				fmt.Fprintf(*conn, text)

			} else if key == 'q' {
				fmt.Println("Quit key sent to server...")
				text := "keyquitTheGame\n"
				fmt.Fprintf(*conn, text)
			} else if key == 'k' {
				fmt.Println("Shuting down all components...")
				text := "keyshutDown\n"
				fmt.Fprintf(*conn, text)
				time.Sleep(3 * time.Second)
				os.Exit(3)
			}

		default:
		}
	}
}

//Send the event to IO in order to save a PGM
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
		//FILTER FOR TCP MESSAGES
		msg, err := reader.ReadString('\n')
		if err != nil || len(msg) == 1 {
			continue
		}
		//If the message contains the matrix to be save
		if msg[1] == 'm' && msg[2] == 'a' && msg[3] == 'p' {
			p := stringToMatrix(msg)
			saveTheWorld(c, p)

			//If the message is related to pressed keys
		} else if msg[0] == 'k' && msg[1] == 'e' && msg[2] == 'y' {

			var state State = 0
			if msg[3] == 'p' {
				state = Paused
			} else if msg[3] == 'e' {
				state = Executing
			} else if msg[3] == 'q' {
				state = Quitting
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
			////If the message contains data for the AliveCellsCount event
		} else if msg[0] == 'a' && msg[1] == 'c' && msg[2] == 'c' {
			i := 3
			turn := 0
			howManyAreAlive := 0
			for i < len(msg) && msg[i] != ' ' {
				turn = turn*10 + (int(msg[i]) - '0')
				i++
			}
			i++
			for i < len(msg) && msg[i] != '\n' {
				howManyAreAlive = howManyAreAlive*10 + (int(msg[i]) - '0')
				i++
			}
			c.events <- AliveCellsCount{
				turn,
				howManyAreAlive,
			}
			//If the message contains data for the CellFlipped event
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
			//If the message reports a TurnComplete event
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
			//If the message reports a FinalTurnComplete event + how many cells are alive
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

//Convert matrix and parameters to one string in order to be sent through TCP
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

//Parse the incoming string to the create computable variables
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

	conn, _ := net.Dial("tcp", "34.230.77.219:8080")
	//LOCALHOST
	//conn, _ := net.Dial("tcp", "127.0.0.1:8080")
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
			if val != 0 {
				c.events <- CellFlipped{
					0,
					util.Cell{
						X: j,
						Y: i,
					},
				}
			}
		}
	}
	//Send Initial board to the engine
	text := convertToString(world, p)
	fmt.Fprintf(conn, text)
	//Start a goroutine for keypresses
	keyTurn := make(chan string)
	go sendKeys(c, &conn, keyTurn)
	//Listen to all incoming messages
	read(c, &conn, keyTurn)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	//c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
