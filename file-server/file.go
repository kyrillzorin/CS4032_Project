package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

func initDB() error {
	var err error
	db, err = bolt.Open("file-server.db", 0600, nil)
	return err
}

func closeDB() {
	db.Close()
}

func ReadFile(filename []byte) []byte {
	var file []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("files"))
		file = b.Get(filename)
		return nil
	})
	return file
}

func WriteFile(filename []byte, filedata []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("files"))
		if err != nil {
			return err
		}
		return b.Put(filename, filedata)
	})
}

func handleClient(message string, conn net.Conn, connReader *bufio.Reader) {
	if strings.HasPrefix(message, "Read") {
		handleRead(message, conn, connReader)
	} else if strings.HasPrefix(message, "Write") {
		handleWrite(message, conn, connReader)
	} else {
		handleDefault(message, conn)
	}
}

func handleRead(message string, conn net.Conn, connReader *bufio.Reader) {
	filepath := strings.TrimPrefix(message, "Read")
	filepath = strings.TrimSpace(filepath)
	file := ReadFile([]byte(filepath))
	fmt.Fprintf(conn, "Send "+filepath+" "+strconv.Itoa(len(file))+"\n")
	conn.Write(file)
}

func handleWrite(message string, conn net.Conn, connReader *bufio.Reader) {
	fileinfostring := strings.TrimPrefix(message, "Write")
	fileinfostring = strings.TrimSpace(fileinfostring)
	fileinfo := strings.Split(fileinfostring, " ")
	filepath := fileinfo[0]
	filelength, _ := strconv.Atoi(fileinfo[1])
	filedata := make([]byte, filelength)
	connReader.Read(filedata)
	err := WriteFile([]byte(filepath), filedata)
	if err != nil {
		fmt.Fprintf(conn, "Receive Failed: "+filepath+"\n")
	} else {
		fmt.Fprintf(conn, "Receive Succeeded: "+filepath+"\n")
	}
}

func handleDefault(message string, conn net.Conn) {
	if message != "" {
		fmt.Println("No such command: "+message)
	}
}
