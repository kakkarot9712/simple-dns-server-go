package main

import (
	"encoding/binary"
	"strconv"
	"strings"
)

type recordType uint8

const (
	A recordType = iota + 1
	_
	_
	_
	CNAME
)

type message struct {
	header   [12]byte
	Question []byte
	Answer   []byte
}

// type DNS struct {
// 	name   string
// 	record recordType
// 	msg message
// }

func getLabelSequence(name string) []byte {
	seq := []byte{}
	labels := strings.Split(name, ".")
	for _, l := range labels {
		seq = append(seq, byte(len(l)))
		seq = append(seq, []byte(l)...)
	}
	seq = append(seq, []byte{0x0}...)
	return seq
}

func (h *message) FillHeader(id [2]byte, QR bool, AA bool, RD bool, RA bool, questions uint16, answers uint16, NSCOUNT [2]byte, ARCOUNT [2]byte) {
	header := [12]byte{}
	// 16 bits => 2 bytes for id
	header[0] = id[0]
	header[1] = id[1]

	bits := ""

	if QR {
		bits += "1"
	} else {
		bits += "0"
	}

	// OPCODE
	bits += "0000"

	if AA {
		bits += "1"
	} else {
		bits += "0"
	}

	// TC
	bits += "0"

	if RD {
		bits += "1"
	} else {
		bits += "0"
	}

	b, err := strconv.ParseUint(bits, 2, 8)
	if err != nil {
		panic(err)
	}
	header[2] = byte(b)

	bits = ""

	if RA {
		bits += "1"
	} else {
		bits += "0"
	}

	// Z
	bits += "000"

	// RCODE
	bits += "0000"

	b, err = strconv.ParseUint(bits, 2, 8)
	if err != nil {
		panic(err)
	}

	header[3] = byte(b)

	queBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(queBytes, questions)
	header[4] = queBytes[0]
	header[5] = queBytes[1]

	ansBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(ansBytes, answers)
	header[6] = ansBytes[0]
	header[7] = ansBytes[1]

	header[8] = NSCOUNT[0]
	header[9] = NSCOUNT[1]

	header[10] = ARCOUNT[0]
	header[11] = ARCOUNT[1]

	h.header = header
}

func (h *message) FillQuestion(name string, record recordType) {
	h.Question = append(h.Question, getLabelSequence(name)...)
	h.Question = append(h.Question, []byte{0x0, byte(record), 0x0, 0x1}...)
}

func (m *message) Bytes() []byte {
	msg := []byte{}
	msg = append(msg, m.header[:]...)
	msg = append(msg, m.Question...)
	msg = append(msg, m.Answer...)
	return msg
}

func (m *message) FillAnswer(name string, record recordType, TTL uint32, ipAddress string) {
	m.Answer = append(m.Answer, getLabelSequence(name)...)
	recordBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(recordBytes, uint16(record))
	m.Answer = append(m.Answer, recordBytes...)

	classBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(classBytes, 1)
	m.Answer = append(m.Answer, classBytes...)

	TTLBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(TTLBytes, TTL)
	m.Answer = append(m.Answer, TTLBytes...)

	RDataStr := strings.Split(ipAddress, ".")
	RData := []byte{}
	for _, ip := range RDataStr {
		ipPart, err := strconv.Atoi(ip)
		if err != nil {
			panic(err)
		}
		RData = append(RData, byte(ipPart))
	}
	Length := make([]byte, 2)
	binary.BigEndian.PutUint16(Length, uint16(len(RData)))
	m.Answer = append(m.Answer, Length...)
	m.Answer = append(m.Answer, RData...)
}
