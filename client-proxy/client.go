package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:8000")
	defer conn.Close()
	fmt.Println("Uppercase Echo TCP Client")
	reader := bufio.NewReader(os.Stdin)
	connReader := bufio.NewReader(conn)
	fmt.Print("Please type your message: ")
	message, _ := reader.ReadString('\n')
	fmt.Fprintf(conn, "GET /echo.php?message="+message+"HTTP/1.0\r\n\r\n")
	response := ""
	for !strings.Contains(response, "\\n") {
		response, _ = connReader.ReadString('\n')
	}
	response = strings.TrimSuffix(response, "\\n")
	fmt.Println("Server Response: " + response)
}
