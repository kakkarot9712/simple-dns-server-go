package main

import (
	"fmt"
	"net"
)

func manageResolver(resolver *net.UDPConn, qc chan question, ac chan answer) {
	defer resolver.Close()
	buf := make([]byte, 512)
	for que := range qc {
		msg := message{}
		msg.header = que.Header
		msg.header.Questions = 1
		// Not replying
		msg.header.QR = false
		msg.question = []question{que}
		response := msg.Bytes()

		_, err := resolver.Write(response)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}

		size, err := resolver.Read(buf)
		if err != nil {
			fmt.Println(err)
		}
		if size != 0 {
			header := ParseHeader(buf[:12])
			if header.Questions > 0 {
				_, newPointer := ParseQuestion(buf[12:size], int(header.Questions))
				if header.Answers > 0 {
					answer, _ := ParseAnswer(buf[12:size], newPointer)
					answer[0].Uid = que.Id
					answer[0].ResponseHeader = header
					ac <- answer[0]
				}
			} else {
				ac <- answer{
					Uid:            que.Id,
					ResponseHeader: header,
				}
			}
		}
	}
}
