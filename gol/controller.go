package gol

import (
	"net"
	"fmt"
	"strings"
	"bufio"
	"strconv"
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

func read(conn *net.Conn) {
	reader := bufio.NewReader(*conn)
	msg, _ := reader.ReadString('\n')
	fmt.Println(msg)
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
		data = append (data, "\n")
	}

	data = append(data, "gata!!\t")

	return strings.Join(data, "")
}

// distributor divides the work between workers and interacts with other goroutines.
func controller(p Params, c distributorChannels) {

  	conn, _ := net.Dial("tcp", "127.0.0.1:8080")
	  
//	for {
    //	fmt.Println("Enter text: ")
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

	//fmt.Println(text)
	fmt.Fprintf(conn, text)
	for {
		read(&conn)
	}

	//}


//	turn := 0

	// TODO: Execute all turns of the Game of Life.
	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

//	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}


