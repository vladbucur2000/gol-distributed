package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type myParameters struct {
	clientid    int
	world       [][]byte
	ImageHeight int
	ImageWidth  int
	Turns       int
	Threads     int
}

func mod(x, m int) int {
	return (x + m) % m
}
func calculateNeighbours(ImageHeight, ImageWidth, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, ImageHeight)][mod(x+j, ImageWidth)] != 0 {
					neighbours++
				}
			}
		}
	}
	return neighbours
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
	//fmt.Println("AICI BA HASHu: ", height)
	world := make([][]byte, height)
	for i := range world {
		world[i] = make([]byte, width)
	}
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			world[i][j] = msg[nr] - '0'
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

func createCellFlipped(i int, j int, turn int) string {
	var data []string
	data = append(data, "cf")
	data = append(data, strconv.Itoa(i))
	data = append(data, " ")
	data = append(data, strconv.Itoa(j))
	data = append(data, " ")
	data = append(data, strconv.Itoa(turn))
	data = append(data, "\n")

	return strings.Join(data, "")

}

func createFinalTurnComplete(turn int, world [][]byte, heigth int, width int) string {
	var data []string
	data = append(data, "ftc")
	data = append(data, strconv.Itoa(turn))
	data = append(data, " ")
	for i := 0; i < heigth; i++ {
		for j := 0; j < width; j++ {
			if world[i][j] != 0 {
				data = append(data, "x:")
				data = append(data, strconv.Itoa(j))
				data = append(data, ";")
				data = append(data, "y:")
				data = append(data, strconv.Itoa(i))
				data = append(data, " ")
			}
		}
	}
	data = append(data, "\n")
	return strings.Join(data, "")
}

func createTurnComplete(turn int) string {
	var data []string
	data = append(data, "tc")
	data = append(data, strconv.Itoa(turn))
	data = append(data, "\n")
	return strings.Join(data, "")
}
func createAliveCellsCount(turn, howManyAreAlive int) string {
	var data []string
	data = append(data, "acc")
	data = append(data, strconv.Itoa(turn))
	data = append(data, " ")
	data = append(data, strconv.Itoa(howManyAreAlive))
	data = append(data, "\n")
	return strings.Join(data, "")
}

func getWorkerWorld(r int, world [][]byte, workerHeight int, ImageHeight int, ImageWidth int, thread int) [][]byte {
	workerWorld := make([][]byte, workerHeight+2)
	for i := range workerWorld {
		workerWorld[i] = make([]byte, ImageWidth)
	}
	//fmt.Println("thread:", thread, (threadworkerHeight+p.ImageHeight-1)%p.ImageHeight)
	//fmt.Println("dimensiuni pentru thread %T height: %T la %T", thread, (threadworkerHeight+p.ImageHeight-r-1)%p.ImageHeight, ((thread+1)workerHeight-r)%p.ImageHeight)
	for j := 0; j < ImageWidth; j++ {
		workerWorld[0][j] = world[mod(thread*workerHeight-r-1, ImageHeight)][j]
	}
	for i := 1; i <= workerHeight; i++ {
		for j := 0; j < ImageWidth; j++ {
			workerWorld[i][j] = world[mod(thread*workerHeight+i-1-r, ImageHeight)][j]

		}
	}
	for j := 0; j < ImageWidth; j++ {
		workerWorld[workerHeight+1][j] =
			world[mod((thread+1)*workerHeight-r, ImageHeight)][j]
	}

	return workerWorld
}

func startNode(workerHeight int, turn int, ImageHeight, ImageWidth int, world [][]byte, thread int, outChannel chan byte, conn *net.Conn, clients map[int]net.Conn) {
	nodeWorldString := convertToString(world, workerHeight, ImageWidth, thread, turn, thread)
	// fmt.Println("Client:", thread)
	// fmt.Println(nodeWorldString)
	fmt.Fprintf(clients[thread], nodeWorldString)
	//go worker(world, turn, ImageHeight, ImageWidth, , height, outChannel, inputChannel, conn*net.Conn)
}

func myVisualiseMatrix(world [][]byte, ImageWidth, ImageHeight int) {
	for i := 0; i < ImageHeight; i++ {
		for j := 0; j < ImageWidth; j++ {
			fmt.Print(world[i][j])
		}
		fmt.Println()
	}
}

func numberToString(nr int) string {
	return strconv.Itoa(nr)
}

func convertToString(world [][]byte, ImageHeight, ImageWidth, Turns, Threads int, clientid int) string {
	var data []string

	hs := numberToString(ImageHeight)
	ws := numberToString(ImageWidth)
	turn := numberToString(Turns)
	thread := numberToString(Threads)
	data = append(data, numberToString(clientid))
	data = append(data, "map")
	data = append(data, hs)
	data = append(data, " ")
	data = append(data, ws)
	data = append(data, " ")
	data = append(data, turn)
	data = append(data, " ")
	data = append(data, thread)
	data = append(data, " ")

	for i := 0; i < ImageHeight; i++ {
		for j := 0; j < ImageWidth; j++ {

			if world[i][j] != 0 {
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

func createStateChange(turn int, state string) string {
	var data []string
	data = append(data, "key")
	data = append(data, state)
	data = append(data, numberToString(turn))
	data = append(data, "\n")
	return strings.Join(data, "")
}
func playTheGame(ImageHeight int, ImageWidth int, Turns int, Threads int, world [][]byte, conn *net.Conn, KeyChannel chan string, clients map[int]net.Conn, nodeChan []chan myParameters) {

	workingWorld := make([][]byte, ImageHeight)
	for i := range world {
		workingWorld[i] = make([]byte, ImageWidth)
	}

	for i := 0; i < ImageHeight; i++ {
		for j := 0; j < ImageWidth; j++ {
			val := world[i][j]
			workingWorld[i][j] = val
			if val != 0 {
				msg := createCellFlipped(i, j, 0)
				fmt.Fprintln(*conn, msg)
			}
		}
	}
	ticker := time.NewTicker(2000 * time.Millisecond)
	turn := 0
	for ; turn < Turns; turn++ {

		select {
		//AICI INTRA KEY URILE
		case key := <-KeyChannel:
			if key == "kpauseTheGame\n" {

				worldString := convertToString(workingWorld, ImageHeight, ImageWidth, Turns, Threads, 0)
				fmt.Fprintln(*conn, worldString)

				fmt.Fprintln(*conn, createStateChange(turn, "p"))

				for {

					key2 := <-KeyChannel
					if key2 == "kpauseTheGame\n" {
						fmt.Fprintln(*conn, createStateChange(turn, "e"))
						break
					}
				}
			} else if key == "ksaveTheGame\n" {
				worldString := convertToString(workingWorld, ImageHeight, ImageWidth, Turns, Threads, 0)
				fmt.Fprintln(*conn, worldString)
			} else if key == "kquitTheGame\n" {
				worldString := convertToString(workingWorld, ImageHeight, ImageWidth, Turns, Threads, 0)

				fmt.Fprintln(*conn, worldString)
				fmt.Fprintln(*conn, createStateChange(turn, "q"))
				//(*conn).Close()
				return
			}
		case <-ticker.C:
			howManyAreAlive := 0
			for i := 0; i < ImageHeight; i++ {
				for j := 0; j < ImageWidth; j++ {
					if workingWorld[i][j] != 0 {
						howManyAreAlive++
					}
				}
			}
			msg := createAliveCellsCount(turn, howManyAreAlive)
			fmt.Fprintln(*conn, msg)
		default:
		}

		nodes := 4
		workerHeight := ImageHeight / nodes
		outChannel := make([]chan byte, nodes)

		newWorld := make([][]byte, ImageHeight)

		for i := range workingWorld {
			newWorld[i] = make([]byte, ImageWidth)
		}

		for thread := 0; thread < nodes; thread++ {

			outChannel[thread] = make(chan byte)

			nodeWorldBytes := getWorkerWorld(0, workingWorld, workerHeight, ImageHeight, ImageWidth, thread)
			startNode(workerHeight+2, turn, ImageHeight, ImageWidth, nodeWorldBytes, thread, outChannel[thread], conn, clients)
		}

		// go receiveWorldsAndProcess()

		unifyWorld := make([][]byte, ImageHeight)
		for i := range unifyWorld {
			unifyWorld[i] = make([]byte, ImageWidth)
		}
		for thread := 0; thread < nodes; thread++ {
			unifyWorldHelper := make([][]byte, workerHeight)
			for i := range unifyWorldHelper {
				unifyWorldHelper[i] = make([]byte, ImageWidth)
			}
			p := <-nodeChan[thread]
			//fmt.Println("inecracam asa : ", p.clientid)
			//Te pup vladuc <3
			for i := 0; i < workerHeight; i++ {
				for j := 0; j < ImageWidth; j++ {
					unifyWorldHelper[i][j] = p.world[i][j]
				}
			}
			for i := 0; i < workerHeight; i++ {
				for j := 0; j < ImageWidth; j++ {
					unifyWorld[mod(thread*workerHeight+i, ImageHeight)][j] = unifyWorldHelper[i][j]
				}
			}

		}

		for i := 0; i < ImageHeight; i++ {
			for j := 0; j < ImageWidth; j++ {
				// if unifyWorld[i][j] != workingWorld[i][j] {
				// 	msg := createCellFlipped(i, j, turn)
				// 	fmt.Fprintf(*conn, msg)
				// }
				workingWorld[i][j] = unifyWorld[i][j]
			}
		}
		turnCompleteString := createTurnComplete(turn)
		fmt.Fprintln(*conn, turnCompleteString)
		//myVisualiseMatrix(workingWorld, ImageWidth, ImageHeight)

	}
	worldString := convertToString(workingWorld, ImageHeight, ImageWidth, Turns, Threads, 0)
	fmt.Fprintln(*conn, worldString)
	finalTurnCompleteString := createFinalTurnComplete(turn, workingWorld, ImageHeight, ImageWidth)
	fmt.Fprintln(*conn, finalTurnCompleteString)

}

func handleConnection(conn *net.Conn, inputChannel chan string, KeyChannel chan string) {

	reader := bufio.NewReader(*conn)
	for {
		msg, er := reader.ReadString('\n')

		if er != nil { //controller has been disconnected
			continue
		}

		if msg[0] == 'k' {
			fmt.Println("AM intrat2")
			KeyChannel <- msg
		} else {

			inputChannel <- msg
		}
	}
}

func acceptConns(ln net.Listener, conns chan net.Conn) {
	for {
		conn, er := ln.Accept()
		if er != nil {
			fmt.Println("Error accepting connection")
		}
		conns <- conn
	}
}

func handleNode(cellFlippedTransition chan string, client *net.Conn, clientid int, nodeChan []chan myParameters) {
	reader := bufio.NewReader(*client)
	for {
		msg, er := reader.ReadString('\n')
		if er != nil {
			continue
		}
		if msg[1] == 'm' && msg[2] == 'a' && msg[3] == 'p' {
			p := stringToMatrix(msg)
			nodeChan[clientid] <- p
		} else if msg[0] == 'c' && msg[1] == 'f' {
			cellFlippedTransition <- msg
		}

	}
}

func handleCellFlippedTransitions(conn *net.Conn, cellFlippedTransition chan string) {
	for {
		select {
		case msg := <-cellFlippedTransition:
			fmt.Fprintf(*conn, msg)
		}
	}
}

func main() {
	KeyChannel := make(chan string)
	inputChannel := make(chan string)

	ln, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("Error!")
	}
	//Create a channel for connections
	conns := make(chan net.Conn)
	//Create a mapping IDs to connections
	clients := make(map[int]net.Conn)

	nodeChan := make([]chan myParameters, 4)
	for i := range nodeChan {
		nodeChan[i] = make(chan myParameters)
	}
	cellFlippedTransition := make(chan string)
	go acceptConns(ln, conns)
	n := 0

	for {
		select {
		case conn := <-conns:

			client := conn
			clients[n] = conn

			if n >= 4 {
				go handleConnection(&conn, inputChannel, KeyChannel)
				go handleCellFlippedTransitions(&conn, cellFlippedTransition)
				msg := <-inputChannel
				p := stringToMatrix(msg)
				playTheGame(p.ImageHeight, p.ImageWidth, p.Turns, p.Threads, p.world, &conn, KeyChannel, clients, nodeChan)
			}

			if n <= 3 {
				go handleNode(cellFlippedTransition, &client, n, nodeChan)
			}

			n++

		/*case msg := <-msgs:
		//TODO Deal with a new message
		// Send the message to all clients that aren't the sender

		for i, client := range clients {
			if msg.sender != i {
				fmt.Fprintf(client, msg.message)
			}
		}*/
		default:
		}

	}
	/*
		for {
			conn, er := ln.Accept()

			if er != nil {
				fmt.Println("Error!")
			}

			go handleConnection(&conn, inputChannel, KeyChannel)
			msg := <-inputChannel
			stringToMatrix(msg, &conn, KeyChannel)
			//go handleKeyPresses(&conn, KeyChannel)
		}*/
}
