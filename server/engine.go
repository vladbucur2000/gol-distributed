package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

//Create a message which reports contains the necessary data for the CellFlipped event
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

//Create a message which reports and contains the necessary data for the FinalTurnComplete event
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

//Create a message which reports and contains the necessary data for the TurnComplete event
func createTurnComplete(turn int) string {
	var data []string
	data = append(data, "tc")
	data = append(data, numberToString(turn))
	data = append(data, "\n")
	return strings.Join(data, "")
}

//Create a message which reports and contains the necessary data for the AliveCellsCount event
func createAliveCellsCount(turn, howManyAreAlive int) string {
	var data []string
	data = append(data, "acc")
	data = append(data, strconv.Itoa(turn))
	data = append(data, " ")
	data = append(data, strconv.Itoa(howManyAreAlive))
	data = append(data, "\n")
	return strings.Join(data, "")
}

//Create a message which reports and contains the necessary data for the StateChange event
func createStateChange(turn int, state string) string {
	var data []string
	data = append(data, "key")
	data = append(data, state)
	data = append(data, numberToString(turn))
	data = append(data, "\n")
	return strings.Join(data, "")
}

//Split the board between the AWS Nodes
func getWorkerWorld(r int, world [][]byte, workerHeight int, ImageHeight int, ImageWidth int, thread int) [][]byte {
	workerWorld := make([][]byte, workerHeight+2)
	for i := range workerWorld {
		workerWorld[i] = make([]byte, ImageWidth)
	}
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

//Send the splitted matrix to a specific AWS Node
func startNode(workerHeight int, turn int, ImageHeight, ImageWidth int, world [][]byte, threads int, node int, conn *net.Conn, clients map[int]net.Conn) {
	nodeWorldString := convertToString(world, workerHeight, ImageWidth, turn, threads, node)
	fmt.Fprintf(clients[node], nodeWorldString)
}

//Function to visualise the actual board in terminal
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

//Convert matrix and parameters to one string in order to be sent through TCP
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

//The Manager of Engine
func playTheGame(p myParameters, conn *net.Conn, KeyChannel chan string, clients map[int]net.Conn, nodeChan []chan myParameters) {

	workingWorld := make([][]byte, p.ImageHeight)
	for i := range p.world {
		workingWorld[i] = make([]byte, p.ImageWidth)
	}

	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			val := p.world[i][j]
			workingWorld[i][j] = val
		}
	}
	//Reports every 2 seconds the alive cells
	ticker := time.NewTicker(2000 * time.Millisecond)
	turn := 0
	for ; turn < p.Turns; turn++ {
		select {
		//If any key is pressed
		case key := <-KeyChannel:
			//The Pause key has been pressed
			if key == "keypauseTheGame\n" {
				worldString := convertToString(workingWorld, p.ImageHeight, p.ImageWidth, p.Turns, p.Threads, 0)
				fmt.Fprintln(*conn, worldString)
				fmt.Fprintln(*conn, createStateChange(turn, "p"))
				for {
					//The Pause key has been pressed again, so start executing
					key2 := <-KeyChannel
					if key2 == "keypauseTheGame\n" {
						fmt.Fprintln(*conn, createStateChange(turn, "e"))
						break
					}
				}
				//The Save key has been pressed
			} else if key == "keysaveTheGame\n" {
				//Send the actual board to the controller in order to be saved as a PGM
				worldString := convertToString(workingWorld, p.ImageHeight, p.ImageWidth, p.Turns, p.Threads, 0)
				fmt.Fprintln(*conn, worldString)
			} else if key == "keyquitTheGame\n" {
				//The Quit key has been pressed -> The actual board will be saved -> the controller disconnects
				worldString := convertToString(workingWorld, p.ImageHeight, p.ImageWidth, p.Turns, p.Threads, 0)
				fmt.Fprintln(*conn, worldString)
				fmt.Fprintln(*conn, createStateChange(turn, "q"))
				return
			} else if key == "keyshutDown\n" {
				//The ShutDown key has been pressed -> All the distributed components will shut down cleanly
				worldString := convertToString(workingWorld, p.ImageHeight, p.ImageWidth, p.Turns, p.Threads, 0)
				fmt.Fprintln(*conn, worldString)
				fmt.Fprintln(*conn, createStateChange(turn, "q"))
				msg := "keyshutDown\n"
				for i := 0; i < 4; i++ {
					fmt.Fprintf(clients[i], msg)
				}
				os.Exit(3)
			}
		case <-ticker.C:
			//AliveCellsCount every 2 seconds
			howManyAreAlive := 0
			for i := 0; i < p.ImageHeight; i++ {
				for j := 0; j < p.ImageWidth; j++ {
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
		workerHeight := p.ImageHeight / nodes

		//Split and send the world between nodes in order to start computing
		for node := 0; node < nodes; node++ {
			nodeWorldBytes := getWorkerWorld(0, workingWorld, workerHeight, p.ImageHeight, p.ImageWidth, node)
			startNode(workerHeight+2, turn, p.ImageHeight, p.ImageWidth, nodeWorldBytes, p.Threads, node, conn, clients)
		}
		//Auxiliary world
		unifyWorld := make([][]byte, p.ImageHeight)
		for i := range unifyWorld {
			unifyWorld[i] = make([]byte, p.ImageWidth)
		}
		//Receive computer worlds from AWS Nodes and start unifying them
		for node := 0; node < nodes; node++ {
			unifyWorldHelper := make([][]byte, workerHeight)
			for i := range unifyWorldHelper {
				unifyWorldHelper[i] = make([]byte, p.ImageWidth)
			}
			params := <-nodeChan[node]
			for i := 0; i < workerHeight; i++ {
				for j := 0; j < params.ImageWidth; j++ {
					unifyWorldHelper[i][j] = params.world[i][j]
				}
			}
			for i := 0; i < workerHeight; i++ {
				for j := 0; j < p.ImageWidth; j++ {
					unifyWorld[mod(node*workerHeight+i, p.ImageHeight)][j] = unifyWorldHelper[i][j]
				}
			}
		}
		//Rewrite the original World and send CellFlipped messages to controller where the cells change state
		for i := 0; i < p.ImageHeight; i++ {
			for j := 0; j < p.ImageWidth; j++ {
				if workingWorld[i][j] != unifyWorld[i][j] {
					msg := createCellFlipped(i, j, turn)
					fmt.Fprintf(*conn, msg)
				}
				workingWorld[i][j] = unifyWorld[i][j]
			}
		}
		//Reports to controller that a turn has been completed
		turnCompleteString := createTurnComplete(turn)
		fmt.Fprintln(*conn, turnCompleteString)
		myVisualiseMatrix(workingWorld, p.ImageWidth, p.ImageHeight)

	}
	//Reports to controller that all turns have been completed and send the last board as well
	worldString := convertToString(workingWorld, p.ImageHeight, p.ImageWidth, p.Turns, p.Threads, 0)
	fmt.Fprintln(*conn, worldString)
	finalTurnCompleteString := createFinalTurnComplete(turn, workingWorld, p.ImageHeight, p.ImageWidth)
	fmt.Fprintln(*conn, finalTurnCompleteString)

}

//Handle connection between controller and engine
func handleConnection(conn *net.Conn, inputChannel chan string, KeyChannel chan string) {
	reader := bufio.NewReader(*conn)
	for {
		msg, er := reader.ReadString('\n')
		if er != nil {
			continue
		}
		if msg[0] == 'k' {
			//Receive key
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

//Handle connection between AWS Nodes and the Engine
func handleNode(client *net.Conn, clientid int, nodeChan []chan myParameters) {
	reader := bufio.NewReader(*client)
	for {
		msg, er := reader.ReadString('\n')
		if er != nil {
			continue
		}
		if msg[1] == 'm' && msg[2] == 'a' && msg[3] == 'p' {
			//Receive computer board from nodes
			p := stringToMatrix(msg)
			nodeChan[clientid] <- p
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
	go acceptConns(ln, conns)
	n := 0

	for {
		//FILTER CONNECTIONS
		select {
		case conn := <-conns:

			client := conn
			clients[n] = conn

			if n >= 4 {
				//Connect the controller -> Parse the board -> start the game
				go handleConnection(&conn, inputChannel, KeyChannel)
				msg := <-inputChannel
				p := stringToMatrix(msg)
				playTheGame(p, &conn, KeyChannel, clients, nodeChan)
			}
			//The first 3 connections will be AWS Nodes
			if n <= 3 {
				go handleNode(&client, n, nodeChan)
			}

			n++
		default:
		}

	}
}
