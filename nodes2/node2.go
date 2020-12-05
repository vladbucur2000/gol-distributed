package main

import (
	"net"
	"fmt"
	"bufio"
)

func handleConnection (conn *net.Conn) {
	
	reader := bufio.NewReader(*conn)

	for {
		msg, _ := reader.ReadString('\n')
		fmt.Printf(msg)
		//fmt.Fprintln(*conn, "OK")
	}
	
  }



func main () {
	ln, err := net.Listen("tcp", ":8889")

	if (err != nil) {
		fmt.Println("Error!")
	}

	for {
		conn, err := ln.Accept()

		if (err != nil) {
		  fmt.Println("Error!")
		}
	
		go handleConnection(&conn)
	}
}