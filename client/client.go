package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"io/ioutil"

	"github.com/kyrillzorin/CS4032_Project/clientproxy"
)﻿

func init() {
	err = clientproxy.Init()
	if err != nil {
		fmt.Println("Error opening DB:", err.Error())
		os.Exit(1)
	}
}

func main() {
	defer clientproxy.closeDB()
	file := clientProxy.Open(test.jpg)
	data, _ := ioutil.ReadFile("./test.jpeg")
	file.Write(data)
	file.Close()
	
	file = clientProxy.Open(test.jpg)
	data = file.ReadByte()
	ioutil.WriteFile("output.jpg", data, 0644)
}
