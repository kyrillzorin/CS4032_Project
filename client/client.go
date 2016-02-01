package main

import (
	"fmt"
	"os"
	"io/ioutil"

	"github.com/kyrillzorin/CS4032_Project/clientproxy"
)

func init() {
	err := clientproxy.Init("localhost:8001")
	if err != nil {
		fmt.Println("Error opening DB:", err.Error())
		os.Exit(1)
	}
}

func main() {
	defer clientproxy.CloseDB()
	file, _ := clientproxy.Open("test.jpg")
	data, _ := ioutil.ReadFile("./test.jpeg")
	file.Write(data)
	file.Close()
	
	file, _ = clientproxy.Open("test.jpg")
	data, _ = file.ReadByte()
	ioutil.WriteFile("output.jpg", data, 0644)
	file.Close()
}
