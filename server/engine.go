package main

import (
	"fmt"
	"net"
	"bufio"
)

func handleConnection(conn *net.Conn) {
	reader := bufio.NewReader(*conn)

	for {
		msg, _ := reader.ReadString('\n')
		fmt.Printf(msg)
		fmt.Fprintln(*conn, "am primit cumetre")
	}
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
