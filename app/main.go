package main

import (
	"fmt"
	"net"
)

// Ensures gofmt doesn't remove the "net" import in stage 1 (feel free to remove this!)
var _ = net.ListenUDP

func main() {
	// Uncomment this block to pass the first stage
	//
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)
		if size == 0 {
			continue
		}

		msg := message{}
		// 1234 Big-endian
		id := [2]byte{0x04, 0xD2}
		QR := true
		AA := false
		RD := false
		RA := false
		var questions uint16 = 1
		var answers uint16 = 1
		NSCOUNT := [2]byte{0x0, 0x0}
		ARCOUNT := [2]byte{0x0, 0x0}
		msg.FillHeader(id, QR, AA, RD, RA, questions, answers, NSCOUNT, ARCOUNT)

		domainName := "codecrafters.io"
		record := A
		msg.FillQuestion(domainName, record)

		ipAddress := "8.8.8.8"
		msg.FillAnswer(domainName, record, 60, ipAddress)

		response := msg.Bytes()

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
