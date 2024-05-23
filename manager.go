package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

func handleStream(stream network.Stream, host host.Host) {
	fmt.Println("Got a new stream!")

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw, host)
	go writeData(rw)
}

func readData(rw *bufio.ReadWriter, host host.Host) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer:", err)
			return
		}

		if str != "\n" {
			currentTime := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("%s\n", currentTime)
			fmt.Printf("From ID : %s\n", host.ID())
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin:", err)
			return
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer:", err)
			return
		}

		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer:", err)
			return
		}
	}
}
