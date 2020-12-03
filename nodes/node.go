package main

import (
	"fmt"
	"net"
	"bufio"
)

func read(conn *net.Conn) {
	reader := bufio.NewReader(*conn)
	for {
		line, _ := reader.ReadString('\n')
		fmt.Println(":: ", line)
	}
}

func write(conn *net.Conn) {
	for {
		fmt.Fprintf(*conn, "sati trag la muie\n")
	}
}

func main() {

	conn, _ := net.Dial("tcp", "127.0.0.1:8080")

	go read(&conn)
	write(&conn)
	
}
