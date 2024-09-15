package main

import (
	"fmt"
	"net"
	"os"
	"slices"
)

func main() {
	resolverIndex := slices.Index(os.Args, "--resolver") + 1
	resolverHost := os.Args[resolverIndex]
	qc := make(chan question)
	ac := make(chan answer)
	// Try to connect to resolver first
	reolverUdp, err := net.ResolveUDPAddr("udp", resolverHost)
	if err != nil {
		panic("Resolver host resolution failed! server can't start: " + err.Error())
	}
	resolver, err := net.DialUDP("udp", nil, reolverUdp)
	if err != nil {
		panic("Connection to resolver server failed! server can't start: " + err.Error())
	}
	go manageResolver(resolver, qc, ac)

	// Now start the server
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
	DNSMappinng := map[int]DNSMessage{}
	buf := make([]byte, 512)
	answers := map[int][]answer{}
	for {
		go ReadFromUDPConn(udpConn, &buf, &DNSMappinng, qc)
		select {
		case ans := <-ac:
			answers[ans.Uid] = append(answers[ans.Uid], ans)
			dns := DNSMappinng[ans.Uid]
			if ans.ResponseHeader.OPCODE != "0000" {
				// Not Implemented
				msg := message{}
				msg.header = header{
					Id:        ans.ResponseHeader.Id,
					QR:        true,
					Questions: ans.ResponseHeader.Questions,
					OPCODE:    ans.ResponseHeader.OPCODE,
					RCODE:     ans.ResponseHeader.RCODE,
					RD:        dns.header.RD,
				}
				msg.question = dns.questions
				response := msg.Bytes()
				go SendAnswersToSource(udpConn, response, dns.source)
				delete(answers, ans.Uid)
				delete(DNSMappinng, ans.Uid)
			} else if len(answers[ans.Uid]) == len(dns.questions) {
				msg := message{}
				msg.header = header{
					Id:        dns.header.Id,
					Questions: uint16(len(dns.questions)),
					Answers:   uint16(len(answers[ans.Uid])),
					QR:        true,
					OPCODE:    dns.header.OPCODE,
					RD:        dns.header.RD,
				}
				msg.question = dns.questions
				msg.Answer = answers[ans.Uid]
				source := dns.source
				response := msg.Bytes()
				go SendAnswersToSource(udpConn, response, source)
				delete(answers, ans.Uid)
				delete(DNSMappinng, ans.Uid)
			}
		default:
		}
	}
}
