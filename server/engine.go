package main

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"strconv"
)


func stringToMatrix (msg string, conn *net.Conn) {
	i := 0
	height := 0
	width := 0
	turn := 0
	thread := 0

	for (i < len(msg) && msg[i] != '\n') {
		height = height * 10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for (i < len(msg) && msg[i] != '\n') {
		width = width * 10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for (i < len(msg) && msg[i] != '\n') {
		turn = turn * 10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for (i < len(msg) && msg[i] != '\n') {
		thread = thread * 10 + (int(msg[i]) - '0')
		i++
	}
	fmt.Println("height:", height)
	fmt.Println("width:", width)
	fmt.Println("turn: ", turn)
	fmt.Println("thread: ", thread)

	i++
	nr := i

	world := make([][]byte, height)
	for i := range world {
		world[i] = make([]byte, width)
	  }

	for i := 0; i < 16; i++ {
	  for j := 0; j < 16; j++ {
		  world[i][j] = msg[nr] - '0'
		  nr++
		  fmt.Print(world[i][j])
		  fmt.Print (" ")
		  
	  }
	  fmt.Println()
	  nr++
	}

	makeTurn(height, width, turn, thread, world, conn)
//	fmt.Println("gata cumetre")


}


func createCellFlipped(i int, j int, turn int) string {
	var data []string
	data = append(data, strconv.Itoa(i))
	data = append(data, " ")
	data = append(data, strconv.Itoa(j))
	data = append(data, " ")
	data = append(data, strconv.Itoa(turn))

	return strings.Join(data, "")

}

func makeTurn (height int, width int, turn int, thread int, world [][]byte, conn *net.Conn) {
	currentTurn := 0
	str := ""
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if (world[i][j] == 1) {
			//	SEND EVENT TO CONTROLLER
				str = createCellFlipped(i, j, currentTurn)
				fmt.Println("salut", str)
				fmt.Fprintln(*conn, str)
			}
		}
	}

}


func handleConnection(conn *net.Conn) {
	reader := bufio.NewReader(*conn)

	//for {
		msg, _ := reader.ReadString('\t')
	//	fmt.Printf(msg)
		stringToMatrix(msg, conn) 
	//	fmt.Printf("\n\n\n")
		fmt.Fprintln(*conn, "am primit cumetre")
//	}
}


func main() {
	ln, err := net.Listen("tcp", ":8080")

	if (err != nil) {
		fmt.Println("eroare cumetre")
	}

	for {
		conn, er := ln.Accept();

		if (er != nil) {
			fmt.Println("eroare cumetre2")
		}

		go handleConnection(&conn)

	}
}
