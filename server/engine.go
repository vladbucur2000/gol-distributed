package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

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

func stringToMatrix(msg string, conn *net.Conn, KeyChannel chan string) {
	i := 0
	height := 0
	width := 0
	turn := 0
	thread := 0

	for i < len(msg) && msg[i] != '\n' {
		height = height*10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for i < len(msg) && msg[i] != '\n' {
		width = width*10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for i < len(msg) && msg[i] != '\n' {
		turn = turn*10 + (int(msg[i]) - '0')
		i++
	}
	i++
	for i < len(msg) && msg[i] != '\n' {
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

	playTheGame(height, width, turn, thread, world, conn, KeyChannel)

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
func worker(originalWorld [][]byte, turn int, ImageHeight, ImageWidth int, start int, end int, height int, outChannel chan byte, inputChannel chan byte, conn *net.Conn) {
	world := make([][]byte, ImageHeight)
	for i := range world {
		world[i] = make([]byte, ImageWidth)
	}
	for i := start; i < end; i++ {
		for j := 0; j < ImageWidth; j++ {
			world[i][j] = <-inputChannel
		}
	}

	newWorld := make([][]byte, ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, ImageWidth)
	}

	for i := start; i < end; i++ {
		for j := 0; j < ImageWidth; j++ {
			neighbours := calculateNeighbours(ImageHeight, ImageWidth, j, i, originalWorld)
			if world[i][j] != 0 {
				if neighbours == 2 || neighbours == 3 {
					newWorld[i][j] = 1
				} else {
					newWorld[i][j] = 0
					msg := createCellFlipped(i, j, turn)
					fmt.Fprintln(*conn, msg)
				}
			} else {
				if neighbours == 3 {
					newWorld[i][j] = 1
					msg := createCellFlipped(i, j, turn)
					fmt.Fprintln(*conn, msg)
				} else {
					newWorld[i][j] = 0
				}
			}
		}

	}

	for i := start; i < end; i++ {
		for j := 0; j < ImageWidth; j++ {
			outChannel <- newWorld[i][j]
		}
	}

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
			fmt.Print(mod(thread*workerHeight+i-1-r, ImageHeight))
			fmt.Print("/")
		}
	}
	for j := 0; j < ImageWidth; j++ {
		workerWorld[workerHeight+1][j] =
			world[mod((thread+1)*workerHeight-r, ImageHeight)][j]
	}

	return workerWorld
}

func startNode(workerHeight int, turn int, ImageHeight, ImageWidth int, world [][]byte, thread int, outChannel chan byte, conn *net.Conn) {
	nodeWorldString := convertToString(world, ImageHeight, ImageWidth, 0, 0)
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

func convertToString(world [][]byte, ImageHeight, ImageWidth, Turns, Threads int) string {
	var data []string

	hs := numberToString(ImageHeight)
	ws := numberToString(ImageWidth)
	turn := numberToString(Turns)
	thread := numberToString(Threads)

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
func playTheGame(ImageHeight int, ImageWidth int, Turns int, Threads int, world [][]byte, conn *net.Conn, KeyChannel chan string) {

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
			if key == "kpauseTheGame\t" {

				worldString := convertToString(workingWorld, ImageHeight, ImageWidth, Turns, Threads)
				fmt.Fprintln(*conn, worldString)

				fmt.Fprintln(*conn, createStateChange(turn, "p"))
				fmt.Println("AM intrat2 da dupa alea")
				for {

					key2 := <-KeyChannel
					if key2 == "kpauseTheGame\t" {
						fmt.Fprintln(*conn, createStateChange(turn, "e"))
						break
					}
				}
			} else if key == "ksaveTheGame\t" {
				worldString := convertToString(workingWorld, ImageHeight, ImageWidth, Turns, Threads)
				fmt.Fprintln(*conn, worldString)
			} else if key == "kquitTheGame\t" {
				worldString := convertToString(workingWorld, ImageHeight, ImageWidth, Turns, Threads)

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
			startNode(workerHeight, turn, ImageHeight, ImageWidth, nodeWorldBytes, thread, outChannel[thread], conn)
		}

		for thread := 0; thread < nodes; thread++ {

			outWorld := make([][]byte, workerHeight)
			for i := range outWorld {
				outWorld[i] = make([]byte, ImageWidth)
			}
			for i := 0; i < workerHeight; i++ {
				for j := 0; j < ImageWidth; j++ {
					outWorld[i][j] = <-outChannel[thread]
				}
			}
			for i := 0; i < workerHeight; i++ {
				for j := 0; j < ImageWidth; j++ {
					world[(thread*workerHeight+i+ImageHeight)%ImageHeight][j] = outWorld[i][j]
				}
			}
		}

		for i := 0; i < ImageHeight; i++ {
			for j := 0; j < ImageWidth; j++ {
				workingWorld[i][j] = newWorld[i][j]
			}
		}
		turnCompleteString := createTurnComplete(turn)
		fmt.Fprintln(*conn, turnCompleteString)
		myVisualiseMatrix(workingWorld, ImageWidth, ImageHeight)

	}
	worldString := convertToString(workingWorld, ImageHeight, ImageWidth, Turns, Threads)
	fmt.Fprintln(*conn, worldString)
	finalTurnCompleteString := createFinalTurnComplete(turn, workingWorld, ImageHeight, ImageWidth)
	fmt.Fprintln(*conn, finalTurnCompleteString)

}

func handleConnection(conn *net.Conn, inputChannel chan string, KeyChannel chan string) {

	reader := bufio.NewReader(*conn)
	for {
		msg, er := reader.ReadString('\t')

		if er != nil { //controller has been disconnected
			continue
		}

		if msg[0] == 'k' {
			fmt.Println("AM intrat2")
			//fmt.Println(msg)
			KeyChannel <- msg
		} else {
			inputChannel <- msg
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

	for {
		conn, er := ln.Accept()

		if er != nil {
			fmt.Println("Error!")
		}

		go handleConnection(&conn, inputChannel, KeyChannel)
		msg := <-inputChannel
		stringToMatrix(msg, &conn, KeyChannel)
		//go handleKeyPresses(&conn, KeyChannel)
	}
}
