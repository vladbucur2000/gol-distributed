package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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

func worker(conn *net.Conn, wholeImage int, p myParameters, start int, end int, height int, width int, outChannel chan byte, inputChannel chan byte) {

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}
	for i := start; i < end; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			val := <-inputChannel
			world[i][j] = val
		}
	}

	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}

	for i := start; i < end; i++ {
		for j := 0; j < width; j++ {
			neighbours := calculateNeighbours(p.ImageHeight, p.ImageWidth, j, i, p.world)
			if world[i][j] != 0 {
				if neighbours == 2 || neighbours == 3 {
					newWorld[i][j] = 1
				} else {
					newWorld[i][j] = 0
				}
			} else {
				if neighbours == 3 {
					newWorld[i][j] = 1
				} else {
					newWorld[i][j] = 0
				}
			}
		}
	}

	for i := start; i < end; i++ {
		for j := 0; j < width; j++ {
			outChannel <- newWorld[i][j]
		}
	}

}

func numberToString(nr int) string {
	return strconv.Itoa(nr)
}
func convertToString(world [][]byte, p myParameters) string {
	var data []string

	hs := numberToString(p.ImageHeight - 2)
	ws := numberToString(p.ImageWidth)
	turn := numberToString(p.Turns)
	thread := numberToString(p.Threads)

	data = append(data, numberToString(p.clientid))
	data = append(data, "map")
	data = append(data, hs)
	data = append(data, " ")
	data = append(data, ws)
	data = append(data, " ")
	data = append(data, turn)
	data = append(data, " ")
	data = append(data, thread)
	data = append(data, " ")

	for i := 1; i < p.ImageHeight-1; i++ {
		for j := 0; j < p.ImageWidth; j++ {

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

func startWorker(conn *net.Conn, wholeImage int, workerHeight int, p myParameters, start int, end int, thread int, outChannel chan byte, inputChannel chan byte) {
	go worker(conn, wholeImage, p, start, end, workerHeight, p.ImageWidth, outChannel, inputChannel)
	for i := start; i < end; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			inputChannel <- p.world[i][j]
		}
	}
}
func computeWorld(conn *net.Conn, p myParameters) [][]byte {

	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}

	wholeImage := 4 * (p.ImageHeight - 2)
	workerHeight := (p.ImageHeight - 2) / p.Threads
	outChannel := make([]chan byte, p.Threads)

	for thread := 0; thread < p.Threads-1; thread++ {
		inputChannel := make(chan byte)
		outChannel[thread] = make(chan byte)
		start := workerHeight*thread + 1
		end := start + workerHeight
		fmt.Println(start, end)
		startWorker(conn, wholeImage, workerHeight, p, start, end, thread, outChannel[thread], inputChannel)
		for i := start; i < end; i++ {
			for j := 0; j < p.ImageWidth; j++ {
				newWorld[i][j] = <-outChannel[thread]
			}
		}
	}
	inputChannel := make(chan byte)
	outChannel[p.Threads-1] = make(chan byte)
	start := workerHeight*(p.Threads-1) + 1
	end := p.ImageHeight - 1
	fmt.Println(start, end)
	fmt.Println("-------")
	startWorker(conn, wholeImage, workerHeight, p, start, end, p.Threads-1, outChannel[p.Threads-1], inputChannel)
	for i := start; i < end; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			newWorld[i][j] = <-outChannel[p.Threads-1]
		}
	}
	return newWorld
}

func read(conn *net.Conn) {
	reader := bufio.NewReader(*conn)
	//n := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		if line[1] == 'm' && line[2] == 'a' && line[3] == 'p' {
			p := stringToMatrix(line)
			newWorld := computeWorld(conn, p)
			newWorldString := convertToString(newWorld, p)
			//fmt.Println(newWorldString)
			fmt.Fprintf(*conn, newWorldString)
		}
		if line == "kshutDown\n" {
			os.Exit(3)
		}

	}
}

func write(conn *net.Conn) {
	for {
	}
}

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:8080")
	go read(&conn)
	write(&conn)
}
