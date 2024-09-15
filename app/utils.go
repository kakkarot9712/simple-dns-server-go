package main

import (
	"fmt"
	"net"
)

var num = 1

func ReadFromUDPConn(udpConn *net.UDPConn, buf *[]byte, dnsM *map[int]DNSMessage, qc chan question) {
	size, source, err := udpConn.ReadFromUDP(*buf)
	if err != nil {
		fmt.Println("Error receiving data:", err)
	}
	header := ParseHeader((*buf)[:12])

	question, _ := ParseQuestion((*buf)[12:size], int(header.Questions))

	dns := DNSMessage{
		source:    source,
		id:        header.Id,
		questions: question,
		header:    header,
	}
	(*dnsM)[num] = dns
	num++
	for _, que := range question {
		que.Header = header
		que.Id = num - 1
		qc <- que
	}
}

func SendAnswersToSource(udpConn *net.UDPConn, response []byte, source *net.UDPAddr) {
	_, err := udpConn.WriteToUDP(response, source)
	if err != nil {
		fmt.Println("Failed to send response:", err)
	}
}
