package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Player struct {
	conn     net.Conn
	nickname string
	x, y, z  int
}

var players = struct {
	sync.RWMutex
	m map[string]*Player
}{m: make(map[string]*Player)}

func main() {
	port := 4000
	if len(os.Args) > 1 {
		var err error
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatal("Invalid port number")
		}
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	fmt.Printf("Server started on port %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		data := scanner.Text()
		parts := strings.SplitN(data, ":", 2)
		cmd, payload := parts[0], parts[1]

		players.Lock()
		player, ok := players.m[conn.RemoteAddr().String()]
		players.Unlock()

		switch cmd {
		case "JOIN":
			if !ok {
				player = &Player{conn: conn, nickname: payload, x: 0, y: 0, z: 0}
				players.Lock()
				players.m[conn.RemoteAddr().String()] = player
				players.Unlock()
				conn.Write([]byte("WELCOME:" + payload + "\n"))
			} else {
				conn.Write([]byte("ERROR:Already joined\n"))
			}
		case "MOVE":
			if ok {
				coords := strings.Split(payload, ",")
				player.x, _ = strconv.Atoi(coords[0])
				player.y, _ = strconv.Atoi(coords[1])
				player.z, _ = strconv.Atoi(coords[2])
			} else {
				conn.Write([]byte("ERROR:Not joined\n"))
			}
		case "LEAVE":
			if ok {
				players.Lock()
				delete(players.m, conn.RemoteAddr().String())
				players.Unlock()
				conn.Write([]byte("GOODBYE:" + player.nickname + "\n"))
				return
			} else {
				conn.Write([]byte("ERROR:Not joined\n"))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading from connection:", err)
	}

	players.Lock()
	delete(players.m, conn.RemoteAddr().String())
	players.Unlock()
}
