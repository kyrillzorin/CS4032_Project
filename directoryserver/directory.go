package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
)

var db *bolt.DB
var currentServer []byte = nil
var locks = struct{
    sync.RWMutex
    m map[string]bool
}{m: make(map[string]bool)}

func initDB() error {
	var err error
	db, err = bolt.Open("directoryserver.db", 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("servers"))
		tx.CreateBucketIfNotExists([]byte("locations"))
		return nil
	})
	return err
}

func closeDB() {
	db.Close()
}

func getServer(serverID []byte) []byte {
	var server []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		server = b.Get(serverID)
		return nil
	})
	if server == nil {
		server = []byte("")
	}
	return server
}

func setServer(serverID []byte, serverinfo []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("servers"))
		if err != nil {
			return err
		}
		return b.Put(serverID, serverinfo)
	})
}

func selectServer() []byte {
	var server []byte
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("servers")).Cursor()
		var k, v []byte
		if currentServer == nil {
			k, v = c.First()
		} else {
			k, v = c.Seek(currentServer)
		}
		k, v = c.Next()
		if k == nil || v == nil{
			k, v = c.First()
		}
		currentServer = k
		server = k
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
	return getServer(location)
}

func setFileLocation(filename []byte) ([]byte, error) {
	location:= selectServer()
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
	locks.RLock()
	defer locks.RUnlock()
	return []byte(strconv.FormatBool(locks.m[string(filename)]))
}

func setLock(filename []byte, status bool) error {
	locks.Lock()
	locks.m[string(filename)] = status
	locks.Unlock()
	return nil
}

func handleClient(message string, conn net.Conn, connReader *bufio.Reader) {
	if strings.HasPrefix(message, "Open ") {
		handleOpen(message, conn, connReader)
	} else if strings.HasPrefix(message, "Close ") {
		handleClose(message, conn, connReader)
	} else if strings.HasPrefix(message, "RegisterNode ") {
		handleRegister(message, conn, connReader)
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
		fmt.Fprintf(conn, "Location "+filepath+" "+string(location)+"\n")
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

func handleRegister(message string, conn net.Conn, connReader *bufio.Reader) {
	serverinfostring := strings.TrimPrefix(message, "RegisterNode ")
	serverinfostring = strings.TrimSpace(serverinfostring)
	serverinfo := strings.Split(serverinfostring, " ")
	err:= setServer([]byte(serverinfo[0]), []byte(serverinfo[1]))


	if err == nil {
		fmt.Fprintf(conn, "Registered "+serverinfo[0]+"\n")
	} else {
		fmt.Fprintf(conn, "Register Failed: "+serverinfo[0]+"\n")
	}
}

func handleDefault(message string, conn net.Conn) {
	if message != "" {
		fmt.Println("No such command: " + message)
	}
}
