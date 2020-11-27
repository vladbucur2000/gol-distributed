package gol

import (
	"net"
	"fmt"
	"strings"
	"bufio"
	"strconv"
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

func read(conn *net.Conn) {
	reader := bufio.NewReader(*conn)
	msg, _ := reader.ReadString('\n')
	fmt.Println(msg)
}

// distributor divides the work between workers and interacts with other goroutines.
func controller(p Params, c distributorChannels) {

  	conn, _ := net.Dial("tcp", "127.0.0.1:8080")
	  
	for {
    //	fmt.Println("Enter text: ")
    //citesc lumea
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	  }
	
	 /* newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}*/

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

	text := util.MatricesToString(world, nil, p.ImageWidth, p.ImageHeight)
	fmt.Fprintf(conn, text)
    read(&conn)	

}


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


