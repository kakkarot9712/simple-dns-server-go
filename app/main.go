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

		if size == 0 {
			continue
		}

		header := ParseHeader(buf[:12])
		question := ParseQuestion(buf[12:size])

		msg := message{}

		msg.header = Header{
			id:        header.id,
			QR:        true,
			OPCODE:    header.OPCODE,
			AA:        false,
			RD:        header.RD,
			Questions: header.Questions,
			Answers:   2,
			NSCOUNT:   0,
			ARCOunt:   0,
		}

		msg.question = question
		ipAddress := "8.8.8.8"
		for _, q := range question {
			msg.FillAnswer(q.Name, q.Type, 60, ipAddress)
		}

		response := msg.Bytes()
		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
