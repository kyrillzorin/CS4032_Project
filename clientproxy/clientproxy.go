package clientproxy

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"errors"

	"github.com/boltdb/bolt"
)

var db *bolt.DB
var directoryserver string

func Init(server string) error {
	var err error
	db, err = bolt.Open("client.db", 0600, nil)
	directoryserver = server
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("files"))
		tx.CreateBucketIfNotExists([]byte("modified"))
		return nil
	})
	return err
}

func CloseDB() {
	db.Close()
}

type File struct {
	name string
	server string
}

func Open(name string) (*File, error) {
	conn, _ := net.Dial("tcp", directoryserver)
	file := new(File)
	file.name = name
	fmt.Fprintf(conn, "Open "+name+"\n")
	connReader := bufio.NewReader(conn)
	message, _ := connReader.ReadString('\n')
	if strings.HasPrefix(message, "IsLocked ") {
		return file, errors.New("File is locked")
	}
	fileinfostring := strings.TrimPrefix(message, "Location ")
	fileinfostring = strings.TrimSpace(fileinfostring)
	file.server = strings.Split(fileinfostring, " ")[1]
	modified, _ := strconv.Atoi(strings.Split(fileinfostring, " ")[2])
	conn.Close()
	var err error
	if modified > getModified([]byte(name)) {
		conn, _ := net.Dial("tcp", file.server)
		defer conn.Close()
		connReader = bufio.NewReader(conn)
		fmt.Fprintf(conn, "Read "+name+"\n")
		message = ""
		for !strings.HasPrefix(message, "Send ") {
			message, _ = connReader.ReadString('\n')
		}
		filepath := strings.TrimPrefix(message, "Send ")
		filepath = strings.TrimSpace(filepath)
		filedatastring, _ := connReader.ReadString('\n')
		filedatastring = strings.TrimSpace(filedatastring)
		filedata, _ := base64.StdEncoding.DecodeString(filedatastring)
		err = writeFile([]byte(filepath), filedata)
		if err != nil {
			file.name = ""
			return file, err
		}
		setModifiedInt([]byte(name), modified)
	}
	return file, err
}

func (f *File) Close() {
	if f.name != "" {
		conn, _ := net.Dial("tcp", f.server)
		file := readFile([]byte(f.name))
		filebase64 := base64.StdEncoding.EncodeToString(file)
		fmt.Fprintf(conn, "Write "+f.name+"\n")
		fmt.Fprintf(conn, filebase64+"\n")
		connReader := bufio.NewReader(conn)
		message, _ := connReader.ReadString('\n')
		for !strings.HasPrefix(message, "Receive Succeeded: ") {
			if strings.HasPrefix(message, "Receive Failed: ") {
				fmt.Fprintf(conn, "Write "+f.name+"\n")
				fmt.Fprintf(conn, filebase64+"\n")
			}
			message, _ = connReader.ReadString('\n')
		}
		conn.Close()
		conn, _ := net.Dial("tcp", directoryserver)
		defer conn.Close()
		connReader = bufio.NewReader(conn)
		message = ""
		fmt.Fprintf(conn, "Close "+f.name+"\n")
		for !strings.HasPrefix(message, "Unlocked ") {
			if strings.HasPrefix(message, "Unlock Failed: ") {
				fmt.Fprintf(conn, "Close "+f.name+"\n")
			}
			message, _ = connReader.ReadString('\n')
		}
		setModified([]byte(f.name))
		f.name = ""
	}
}

func (f *File) Write(p []byte) (n int, err error) {
	n = 0
	err = errors.New("File is closed")
	if f.name != "" {
		err = writeFile([]byte(f.name), p)
		if err == nil {
			n = len(p)
		}
	}
	return
}

func (f *File) ReadByte() ([]byte, error) {
	err := errors.New("File is closed")
	var data []byte
	if f.name != "" {
		err = nil
		data = readFile([]byte(f.name))
	}
	return data, err
}

func (f *File) Read(p []byte) (n int, err error) {
	n = 0
	err = errors.New("File is closed")
	if f.name != "" {
		filedata := readFile([]byte(f.name))
		n = copy(p, filedata)
		err = nil
	}
	return
}

func readFile(filename []byte) []byte {
	var file []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("files"))
		file = b.Get(filename)
		return nil
	})
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

func getModified(filename []byte) int {
	var modified []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("modified"))
		modified = b.Get(filename)
		return nil
	})
	if modified == nil {
		return 0
	}
	var err error
	modifiedInt, err := strconv.Atoi(modified)
	if err != nil {
		return 0
	}
	return modifiedInt
}

func setModified(filename []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("modified"))
		if err != nil {
			return err
		}
		modified := b.Get(filename)
		if modified == nil {
			modified = strconv.Itoa(0)
		}
		modifiedInt, err := strconv.Atoi(string(modified))
		if err != nil {
			return err
		}
		modifiedInt++
		modified = strconv.Itoa(modifiedInt)
		return b.Put(filename, modified)
	})
	return err
}

func setModifiedInt(filename []byte, data int) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("modified"))
		if err != nil {
			return err
		}
		modified := strconv.Itoa(data)
		return b.Put(filename, modified)
	})
	return err
}
