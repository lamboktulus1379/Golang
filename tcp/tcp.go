package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

/*
 * Complete the 'TCPServer' function below.
 *
 * The function accepts chan bool ready as a parameter.
 */

func TCPServer(ready chan bool) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	ready <- true

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	request(conn)
	response(conn)

	// rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// fmt.Println("Inside handle connection")
	// req, err := rw.Read
	// if err != nil {
	// 	fmt.Println("err", err)
	// 	rw.WriteString("Failed to read string")
	// 	rw.Flush()
	// 	return
	// }
	// fmt.Printf("%s", req)
	// rw.WriteString(fmt.Sprintf("%s", req))
	// rw.Flush()
}

func request(conn net.Conn) {
	var buf [maxBufferSize]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		str := reverse(string(buf[0:n]))

		_, err2 := conn.Write([]byte(str))
		if err2 != nil {
			return
		}
	}
}

func reverse(s string) string {
	rns := []rune(s)
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		rns[i], rns[j] = rns[j], rns[i]
	}
	return string(rns)
}

func response(conn net.Conn) {
	body := `fdafasd`

	fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(body))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	fmt.Fprint(conn, "\r\n")
	fmt.Fprint(conn, body)
}

const maxBufferSize = 1024
const address = "127.0.0.1:3333"

func main() {
	stdout, err := os.Create(os.Getenv("OUTPUT_PATH"))
	checkError(err)

	defer stdout.Close()

	reader := bufio.NewReaderSize(os.Stdin, 16*1024*1024)
	writer := bufio.NewWriterSize(stdout, 16*1024*1024)

	messagesCount, err := strconv.ParseInt(strings.TrimSpace(readLine(reader)), 10, 64)
	checkError(err)

	var messages []string

	for i := 0; i < int(messagesCount); i++ {
		messagesItem := readLine(reader)
		messages = append(messages, messagesItem)
	}

	ready := make(chan bool)
	go TCPServer(ready)
	<-ready
	reversed, err := tcpClient(messages)
	if err != nil {
		panic(err)
	}
	for _, msg := range reversed {
		fmt.Fprintf(writer, "%s\n", msg)
	}
	writer.Flush()
}

// func readLine(reader *bufio.Reader) string {
// 	str, _, err := reader.ReadLine()
// 	if err == io.EOF {
// 		return ""
// 	}

// 	return strings.TrimRight(string(str), "\r\n")
// }

// func checkError(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }

func tcpClient(messages []string) ([]string, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return []string{}, err
	}

	reversed := []string{}

	for _, msg := range messages {

		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return []string{}, err
		}
		_, err = conn.Write([]byte(msg))
		if err != nil {
			return []string{}, err
		}

		reply := make([]byte, maxBufferSize)

		n, err := conn.Read(reply)
		if err != nil {
			return []string{}, err
		}

		reversed = append(reversed, string(reply[:n]))
		conn.Close()
	}

	return reversed, nil
}
