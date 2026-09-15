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
	"time"
)

type storeEntry struct {
	value     string
	expiresAt time.Time // zero value = koi expiry nahi
}

var store = make(map[string]storeEntry)
var mu sync.Mutex

func main() {
	listener, err := net.Listen("tcp", ":6390")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is running on port 6390")
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			sweepExpiredKeys()
		}
	}()
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
			expiresAt := time.Time{}
			if len(command) == 5 && strings.EqualFold(command[3], "PX") {
				ms, err := strconv.Atoi(command[4])
				if err == nil {
					expiresAt = time.Now().Add(time.Duration(ms) * time.Millisecond)
				}
			}
			if len(command) == 5 && strings.EqualFold(command[3], "EX") {
				sec, err := strconv.Atoi(command[4])
				if err == nil {
					expiresAt = time.Now().Add(time.Duration(sec) * time.Second)
				}
			}
			setValue(command[1], command[2], expiresAt)
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
func setValue(key, value string, expiresAt time.Time) {
	mu.Lock()
	defer mu.Unlock()
	store[key] = storeEntry{value: value, expiresAt: expiresAt}
}

func getValue(key string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()
	entry, ok := store[key]
	if !ok {
		return "", false
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		delete(store, key)
		return "", false
	}
	return entry.value, true
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
	entry, ok := store[key]
	if !ok {
		return false
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		delete(store, key)
		return false
	}
	return true
}
func sweepExpiredKeys() {
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	for key, entry := range store {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			delete(store, key)
			fmt.Println("Active expiration removed key:", key)
		}
	}
}
