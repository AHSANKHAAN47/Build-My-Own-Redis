package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var store = make(map[string]string)
var mu sync.Mutex

func main() {
	listener, err := net.Listen("tcp", ":6390")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is running on port 6390")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
func readCommand(reader *bufio.Reader) ([]string, error) {

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(line, "\r\n")
	N, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}
	result := make([]string, N)
	for i := 0; i < N; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSuffix(line, "\r\n")
		L, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		valueBuf := make([]byte, L)
		_, err = io.ReadFull(reader, valueBuf)
		if err != nil {
			return nil, err
		}
		if _, err := reader.Discard(2); err != nil {
			return nil, err
		}
		result[i] = string(valueBuf)
	}
	return result, nil
}
func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Client connected")
	reader := bufio.NewReader(conn)
	for {
		command, err := readCommand(reader)
		if err != nil {
			fmt.Println("Client disconnected:", err)
			return
		}
		fmt.Println("Message received:", command)
		if strings.EqualFold(command[0], "PING") {
			conn.Write([]byte("+PONG\r\n"))
		}
		if strings.EqualFold(command[0], "ECHO") {
			conn.Write([]byte("$" + strconv.Itoa(len(command[1])) + "\r\n" + command[1] + "\r\n"))
		}
		if strings.EqualFold(command[0], "SET") {
			setValue(command[1], command[2])
			conn.Write([]byte("+OK\r\n"))
		}
		if strings.EqualFold(command[0], "GET") {
			value, ok := getValue(command[1])
			if ok {
				conn.Write([]byte("$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"))
			} else {
				conn.Write([]byte("$-1\r\n"))
			}
		}
		if strings.EqualFold(command[0], "DEL") {
			deleted := deleteValue(command[1])
			if deleted {
				conn.Write([]byte(":1\r\n"))
			} else {
				conn.Write([]byte(":0\r\n"))
			}
		}
		if strings.EqualFold(command[0], "EXISTS") {
			exists := existsValue(command[1])
			if exists {
				conn.Write([]byte(":1\r\n"))
			} else {
				conn.Write([]byte(":0\r\n"))
			}
		}
	}
}
func setValue(key, value string) {
	mu.Lock()
	defer mu.Unlock()
	store[key] = value
}

func getValue(key string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()
	value, ok := store[key]
	return value, ok
}
func deleteValue(key string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, ok := store[key]
	delete(store, key)
	return ok
}
func existsValue(key string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, ok := store[key]
	return ok
}
