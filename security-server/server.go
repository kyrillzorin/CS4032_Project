package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http" //used for finding external IP only
	"os"
	"strconv"
	"strings"
)

var (
	IP        = getIP()
	EXT_IP    = getExternalIP()
	PORT      = getPort()
	MaxWorker = getMaxWorkers()
	MaxQueue  = getMaxQueue()
)
var ConnQueue chan net.Conn = make(chan net.Conn, MaxQueue)

type Worker struct {
	WorkerPool  chan chan net.Conn
	ConnChannel chan net.Conn
	quit        chan bool
}

func NewWorker(workerPool chan chan net.Conn) Worker {
	return Worker{
		WorkerPool:  workerPool,
		ConnChannel: make(chan net.Conn),
		quit:        make(chan bool)}
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.ConnChannel

			select {
			case conn := <-w.ConnChannel:
				handleRequest(conn)

			case <-w.quit:
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Supervisor struct {
	WorkerPool chan chan net.Conn
	MaxWorkers int
}

func NewSupervisor(maxWorkers int) *Supervisor {
	pool := make(chan chan net.Conn, maxWorkers)
	return &Supervisor{WorkerPool: pool, MaxWorkers: maxWorkers}
}

func (s *Supervisor) Run() {
	for i := 0; i < s.MaxWorkers; i++ {
		worker := NewWorker(s.WorkerPool)
		worker.Start()
	}

	go s.dispatch()
}

func (s *Supervisor) dispatch() {
	for {
		select {
		case conn := <-ConnQueue:
			go func(conn net.Conn) {
				connChannel := <-s.WorkerPool
				connChannel <- conn
			}(conn)
		}
	}
}

func init() {
	supervisor := NewSupervisor(MaxWorker)
	supervisor.Run()
}

func main() {
	listen, err := net.Listen("tcp", IP+":"+PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listen.Close()
	fmt.Println("Listening on " + IP + ":" + PORT)
	fmt.Println("Max Workers:", MaxWorker)
	fmt.Println("Max Queued Connections:", MaxQueue)
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		ConnQueue <- conn
	}
}

func handleRequest(conn net.Conn) {
	connReader := bufio.NewReader(conn)
	for true {
		message, _ := connReader.ReadString('\n')
		message = strings.TrimSpace(message)
		response := ""
		if strings.HasPrefix(message, "HELO ") {
			text := strings.TrimPrefix(message, "HELO ")
			text = strings.TrimSpace(text)
			response = HELO(text)
		} else if message == "KILL_SERVICE" {
			killService(conn)
		} else {
			handleClient(message, conn, connReader)
		}
		if response != "" {
			fmt.Fprintf(conn, response)
		}
	}
	fmt.Println("lol")
}

func killService(conn net.Conn) {
	conn.Close()
	for _, client := range Clients {
		client.Conn.Close()
	}
	os.Exit(0)
}

func HELO(text string) string {
	return "HELO " + text + "\nIP:" + EXT_IP + "\nPort:" + PORT + "\nStudentID:ea5f6b94d6a8a8f1e7890f6a64883cdc2b6125821e20ddd36a33b773bd46b727\n"
}

func getIP() string {
	e := os.Getenv("CS4032_LAB_4_IP")
	if len(e) > 0 {
		return e
	}
	return "127.0.0.1"
}

func getExternalIP() string {
	e := os.Getenv("CS4032_LAB_4_IP_EXT")
	resp, err := http.Get("http://myexternalip.com/raw")
	if (err != nil) || (e == "false") {
		return IP
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)
	return ip
}

func getPort() string {
	e := os.Getenv("CS4032_LAB_4_PORT")
	if len(e) > 0 {
		return e
	}
	return "8080"
}

func getMaxWorkers() int {
	e := os.Getenv("CS4032_LAB_4_MAX_WORKERS")
	if len(e) > 0 {
		i, _ := strconv.Atoi(e)
		return i
	}
	return 15000
}

func getMaxQueue() int {
	e := os.Getenv("CS4032_LAB_4_MAX_QUEUE")
	if len(e) > 0 {
		i, _ := strconv.Atoi(e)
		return i
	}
	return 20000
}
