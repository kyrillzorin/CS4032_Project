package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

func initDB() error {
	var err error
	db, err = bolt.Open("fileserver.db", 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("files"))
		return nil
	})
	return err
}

func closeDB() {
	db.Close()
}

func readFile(filename []byte) []byte {
	var file []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("files"))
		file = b.Get(filename)
		return nil
	})
	if file == nil {
		file = []byte("")
	}
	return file
}

func writeFile(filename []byte, filedata []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("files"))
		if err != nil {
			return err
		}
		return b.Put(filename, filedata)
	})
}

func handleClient(message string, conn net.Conn, connReader *bufio.Reader) {
	if strings.HasPrefix(message, "Read ") {
		handleRead(message, conn, connReader)
	} else if strings.HasPrefix(message, "Write ") {
		handleWrite(message, conn, connReader)
	} else {
		handleDefault(message, conn)
	}
}

func handleRead(message string, conn net.Conn, connReader *bufio.Reader) {
	filepath := strings.TrimPrefix(message, "Read ")
	filepath = strings.TrimSpace(filepath)
	file := readFile([]byte(filepath))
	filebase64 := base64.StdEncoding.EncodeToString(file)
	fmt.Fprintf(conn, "Send "+filepath+"\n")
	fmt.Fprintf(conn, filebase64+"\n")
}

func handleWrite(message string, conn net.Conn, connReader *bufio.Reader) {
	filepath := strings.TrimPrefix(message, "Write ")
	filepath = strings.TrimSpace(filepath)
	filedatastring, _ := connReader.ReadString('\n')
	filedatastring = strings.TrimSpace(filedatastring)
	filedata, _ := base64.StdEncoding.DecodeString(filedatastring)
	err := writeFile([]byte(filepath), filedata)
	if err != nil {
		fmt.Fprintf(conn, "Receive Failed: "+filepath+"\n")
	} else {
		fmt.Fprintf(conn, "Receive Succeeded: "+filepath+"\n")
	}
}

func handleDefault(message string, conn net.Conn) {
	if message != "" {
		fmt.Println("No such command: " + message)
	}
}
