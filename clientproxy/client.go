package clientproxy

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

func Init() error {
	var err error
	db, err = bolt.Open("client.db", 0600, nil)
	return err
}

func CloseDB() {
	db.Close()
}

type File struct {
	name string
}

func Open(name string) (*File, error) {
	conn, _ := net.Dial("tcp", "127.0.0.1:8000")
	defer conn.Close()
	file := new(File)
	file.name = name
	fmt.Fprintf(conn, "Read "+name+"\n")
	connReader := bufio.NewReader(conn)
	message := ""
	for !strings.HasPrefix(message, "Send ") {
		message, _ := connReader.ReadString('\n')
	}
	fileinfostring := strings.TrimPrefix(message, "Send")
	fileinfostring = strings.TrimSpace(fileinfostring)
	fileinfo := strings.Split(fileinfostring, " ")
	filepath := fileinfo[0]
	filelength, _ := strconv.Atoi(fileinfo[1])
	filedata := make([]byte, filelength)
	connReader.Read(filedata)
	err := writeFile([]byte(filepath), filedata)
	if err != nil {
		file.name = ""
	}
	return file, err
}

func (f *File) Close() {
	conn, _ := net.Dial("tcp", "127.0.0.1:8000")
	defer conn.Close()
	file := readFile([]byte(f.name))
	fmt.Fprintf(conn, "Send "+f.name+" "+strconv.Itoa(len(file))+"\n")
	conn.Write(file)
	connReader := bufio.NewReader(conn)
	message := ""
	message, _ := connReader.ReadString('\n')
	for !strings.HasPrefix(message, "Receive Succeeded: ") {
		if strings.HasPrefix(message, "Receive Failed: ") {
			fmt.Fprintf(conn, "Send "+f.name+" "+strconv.Itoa(len(file))+"\n")
			conn.Write(file)
		}
		message, _ := connReader.ReadString('\n')
	}
	f.name = ""
}

func (f *File) Write(p []byte) (n int, err error) {
	n = 0
	err = writeFile([]byte(f.name), p)
	if err == nil {
		n = len(p)
	}
	return
}

func (f *File) Read() []byte {
	return readFile([]byte(f.name))
}

func (f *File) Read(p []byte) (n int, err error) {
	filedata := readFile([]byte(f.name))
	n = copy(p, filedata)
	err = nil
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
