package main

import (
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
}

func (h *message) FillHeader(id [2]byte, QR bool, AA bool, RD bool, RA bool, questions uint16, ANCOUNT [2]byte, NSCOUNT [2]byte, ARCOUNT [2]byte) {
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

	queBytes := []byte(string(questions))
	if len(queBytes) == 1 {
		header[4] = byte(0x0)
		header[5] = queBytes[0]
	} else {
		header[4] = queBytes[0]
		header[5] = queBytes[1]
	}

	header[6] = ANCOUNT[0]
	header[7] = ANCOUNT[1]

	header[8] = NSCOUNT[0]
	header[9] = NSCOUNT[1]

	header[10] = ARCOUNT[0]
	header[11] = ARCOUNT[1]

	h.header = header
}

func (h *message) FillQuestion(name string, record recordType) {
	labels := strings.Split(name, ".")
	for _, l := range labels {
		h.Question = append(h.Question, byte(len(l)))
		h.Question = append(h.Question, []byte(l)...)
	}
	h.Question = append(h.Question, []byte{0x0}...)
	h.Question = append(h.Question, []byte{0x0, byte(record), 0x0, 0x1}...)
}

func (m *message) Bytes() []byte {
	msg := []byte{}
	msg = append(msg, m.header[:]...)
	msg = append(msg, m.Question...)
	return msg
}
