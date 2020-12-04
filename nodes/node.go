package main

import (
	"bufio"
	"fmt"
	"net"
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

func stringToMatrix(msg string) myParameters {
	clientid := (int(msg[0]) - '0')
	i := 4
	height := 0
	width := 0
	turn := 0
	thread := 0
	//2map14
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

func computeWorld(conn *net.Conn, p myParameters) [][]byte {

	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	wholeImage := 4 * (p.ImageHeight - 2)
	// fmt.Println(mod(p.clientid*(p.ImageHeight-2), wholeImage))
	// fmt.Println(mod(p.clientid*(p.ImageHeight-2)+p.ImageHeight-2, p.ImageHeight))
	for i := 1; i < p.ImageHeight-1; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			neighbours := calculateNeighbours(p.ImageHeight, p.ImageWidth, j, i, p.world)
			if p.world[i][j] != 0 {
				if neighbours == 2 || neighbours == 3 {
					newWorld[i][j] = 1
				} else {
					newWorld[i][j] = 0
					msg := createCellFlipped(mod(p.clientid*(p.ImageHeight-2)+i+wholeImage, wholeImage), j, p.Turns)
					fmt.Fprintf(*conn, msg)
				}
			} else {
				if neighbours == 3 {
					newWorld[i][j] = 1
					msg := createCellFlipped(mod(p.clientid*(p.ImageHeight-2)+i+wholeImage, wholeImage), j, p.Turns)
					fmt.Fprintf(*conn, msg)
				} else {
					newWorld[i][j] = 0
				}
			}
		}

	}

	return newWorld
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

	}
}

func write(conn *net.Conn) {
	for {
		//fmt.Fprintf(*conn, "sati trag la muie\n")
	}
}

func main() {

	conn, _ := net.Dial("tcp", "127.0.0.1:8080")

	go read(&conn)
	write(&conn)

}
