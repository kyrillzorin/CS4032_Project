package main

import (
	"bufio"
	"io/ioutil"
	"fmt"
	"net"
	"strconv"
	"strings"
	"log"

	"github.com/boltdb/bolt"
)

ReadFile(filename []byte) []byte{
	var file []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("files"))
		file = b.Get(filename)
		return nil
	})
	return file
}

WriteFile(filename []byte, filedata []byte) error{
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

func handleRead(message string, conn net.Conn, connReader *bufio.Reader) bool {
	filepath := strings.TrimPrefix(message, "Read")
	filepath = strings.TrimSpace(filepath)
	file = ReadFile([]byte(filepath))
	fmt.Fprintf(conn, "Send " + filepath + "\n")
	fmt.Fprintf(conn, file)
}

func handleWrite(message string, conn net.Conn, connReader *bufio.Reader) bool {
	filepath := strings.TrimPrefix(message, "Write")
	filepath = strings.TrimSpace(filepath)
	// Todo: Read filedata
	err := WriteFile([]byte(filepath), filedata)
	if err != nil {
		fmt.Fprintf(conn, "Receive Failed: " + filepath + "\n")
	} else {
		fmt.Fprintf(conn, "Receive Succeeded: " + filepath + "\n")
	}
}
