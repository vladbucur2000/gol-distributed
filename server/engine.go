package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
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

func stringToMatrix(msg string, conn *net.Conn) {
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

	playTheGame(height, width, turn, thread, world, conn)

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

func startWorker(workerHeight int, turn int, ImageHeight, ImageWidth int, world [][]byte, start int, end int, thread int, outChannel chan byte, inputChannel chan byte, conn *net.Conn) {

	go worker(world, turn, ImageHeight, ImageWidth, start, end, workerHeight, outChannel, inputChannel, conn)

	for i := start; i < end; i++ {
		for j := 0; j < ImageWidth; j++ {

			inputChannel <- world[i][j]
		}

	}
}
func myVisualiseMatrix(world [][]byte, ImageWidth, ImageHeight int) {
	for i := 0; i < ImageHeight; i++ {
		for j := 0; j < ImageWidth; j++ {
			fmt.Print(world[i][j])
		}
		fmt.Println()
	}
}

func playTheGame(ImageHeight int, ImageWidth int, Turns int, Threads int, world [][]byte, conn *net.Conn) {

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

	turn := 0
	for ; turn < Turns; turn++ {

		// select {
		// //AICI INTRA KEY URILE
		// default:
		// }

		workerHeight := ImageHeight / Threads
		outChannel := make([]chan byte, Threads)

		newWorld := make([][]byte, ImageHeight)
		for i := range workingWorld {
			newWorld[i] = make([]byte, ImageWidth)
		}
		for thread := 0; thread < Threads-1; thread++ {
			inputChannel := make(chan byte)
			outChannel[thread] = make(chan byte)
			start := workerHeight * thread
			end := start + workerHeight
			startWorker(workerHeight, turn, ImageHeight, ImageWidth, workingWorld, start, end, thread, outChannel[thread], inputChannel, conn)
			for i := start; i < end; i++ {
				for j := 0; j < ImageWidth; j++ {
					newWorld[i][j] = <-outChannel[thread]
				}
			}
		}

		inputChannel := make(chan byte)
		outChannel[Threads-1] = make(chan byte)
		start := workerHeight * (Threads - 1)
		end := ImageHeight

		startWorker(workerHeight, turn, ImageHeight, ImageWidth, workingWorld, start, end, Threads-1, outChannel[Threads-1], inputChannel, conn)
		for i := start; i < end; i++ {
			for j := 0; j < ImageWidth; j++ {
				newWorld[i][j] = <-outChannel[Threads-1]
			}
		}

		for i := 0; i < ImageHeight; i++ {
			for j := 0; j < ImageWidth; j++ {
				workingWorld[i][j] = newWorld[i][j]
			}
		}
		turnCompleteString := createTurnComplete(turn)

		fmt.Fprintln(*conn, turnCompleteString)
		//myVisualiseMatrix(workingWorld, ImageWidth, ImageHeight)

	}

	finalTurnCompleteString := createFinalTurnComplete(turn, workingWorld, ImageHeight, ImageWidth)
	fmt.Fprintln(*conn, finalTurnCompleteString)

}

func handleConnection(conn *net.Conn) {
	reader := bufio.NewReader(*conn)

	msg, _ := reader.ReadString('\t')

	stringToMatrix(msg, conn)

	fmt.Fprintln(*conn, "am primit cumetre")

}

func main() {
	ln, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("eroare cumetre")
	}

	for {
		conn, er := ln.Accept()

		if er != nil {
			fmt.Println("eroare cumetre2")
		}

		go handleConnection(&conn)

	}
}
