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
var currentServer []byte = nil

func initDB() error {
	var err error
	db, err = bolt.Open("directoryserver.db", 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("servers"))
		tx.CreateBucketIfNotExists([]byte("locations"))
		tx.CreateBucketIfNotExists([]byte("locks"))
		return nil
	})
	return err
}

func closeDB() {
	db.Close()
}

func getServer() []byte {
	var server []byte
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("servers")).Cursor()
		var k, v []byte
		if currentServer == nil {
			k, v := c.First()
		} else {
			k, v := c.Seek(currentServer)
		}
		k, v = c.Next()
		if k == nil || v == nil{
			k, v := c.First()
		}
		currentServer = k
		server = v
		return nil
	})
	return server
}

func getFileLocation(filename []byte) []byte {
	var location []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("locations"))
		location = b.Get(filename)
		return nil
	})
	if location == nil {
		location, _ = setFileLocation(filename)
	}
	return location
}

func setFileLocation(filename []byte) []byte, error {
	location:= getServer()
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("locations"))
		if err != nil {
			return err
		}
		return b.Put(filename, location)
	})
	return location, err
}

func getLock(filename []byte) []byte {
	var lock []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("locks"))
		location = b.Get(filename)
		return nil
	})
	if lock == nil {
		lock = []byte("false")
	}
	return lock
}

func setLock(filename []byte, bool status) error {
	var lock []byte
	if status {
		lock = []byte("true")
	} else {
		lock = []byte("false")
	}
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("locks"))
		if err != nil {
			return err
		}
		return b.Put(filename, lock)
	})
	return err
}

func handleClient(message string, conn net.Conn, connReader *bufio.Reader) {
	if strings.HasPrefix(message, "Open ") {
		handleOpen(message, conn, connReader)
	} else if strings.HasPrefix(message, "Close ") {
		handleClose(message, conn, connReader)
	} else {
		handleDefault(message, conn)
	}
}

func handleOpen(message string, conn net.Conn, connReader *bufio.Reader) {
	filepath := strings.TrimPrefix(message, "Open ")
	filepath = strings.TrimSpace(filepath)
	location := getFileLocation([]byte(filepath))
	lock := string(getLock([]byte(filepath)))
	if lock == "false" {
		setLock([]byte(filepath), true)
		fmt.Fprintf(conn, "Location "+filepath+" "+location+"\n")
	} else {
		fmt.Fprintf(conn, "IsLocked "+filepath+"\n")
	}
}

func handleClose(message string, conn net.Conn, connReader *bufio.Reader) {
	filepath := strings.TrimPrefix(message, "Close ")
	filepath = strings.TrimSpace(filepath)
	err := setLock([]byte(filepath), false)
	if err != nil {
		fmt.Fprintf(conn, "Unlock Failed: "+filepath+"\n")
	} else {
		fmt.Fprintf(conn, "Unlocked "+filepath+"\n")
	}
}

func handleDefault(message string, conn net.Conn) {
	if message != "" {
		fmt.Println("No such command: " + message)
	}
}
