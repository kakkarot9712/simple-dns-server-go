package main

import (
	"encoding/binary"
	"fmt"
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
	header   Header
	question []Question
	Answer   []byte
}

// type DNS struct {
// 	name   string
// 	record recordType
// 	msg message
// }

type Header struct {
	id        uint16
	QR        bool
	OPCODE    string
	AA        bool
	RD        bool
	RA        bool
	RCODE     string
	Questions uint16
	Answers   uint16
	NSCOUNT   uint16
	ARCOunt   uint16
}

type Question struct {
	Name  string
	Type  recordType
	Class uint16
}

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

func getOctetFromByte(b byte) string {
	bits := fmt.Sprintf("%b", b)
	for range 8 - len(bits) {
		bits = "0" + bits
	}
	return bits
}

func ParseHeader(buff []byte) Header {
	h := Header{}
	if len(buff) < 12 {
		fmt.Println("Invalid Header passed!")
		return h
	}
	h.id = binary.BigEndian.Uint16(buff[:2])

	octet := getOctetFromByte(buff[2])

	h.QR = buff[2]&255 == 1
	h.OPCODE = octet[1:5]
	h.AA = buff[2]&4 == 1
	h.RD = buff[2]&1 == 1

	octet = getOctetFromByte(buff[3])
	h.RA = buff[3]&255 == 1
	h.RCODE = octet[4:]

	h.Questions = binary.BigEndian.Uint16(buff[4:6])
	h.Answers = binary.BigEndian.Uint16(buff[6:8])
	h.NSCOUNT = binary.BigEndian.Uint16(buff[8:10])
	h.ARCOunt = binary.BigEndian.Uint16(buff[10:12])
	return h
}

func ParseQuestion(buff []byte) []Question {
	labels := []string{}
	questions := []Question{}
	pointer := 0
	oldPointer := 0
	for {
		b := getOctetFromByte(buff[pointer])
		labelLength := int(buff[pointer])
		pointer++
		labelBytes := []byte{}
		// labels = append(labels, string(buff[pointer:pointer+labelLength]))
		legthProcessed := 0
		for {
			if b[:2] != "11" {
				labelBytes = append(labelBytes, buff[pointer])
				pointer++
				legthProcessed++
			} else {
				bits := b[2:] + getOctetFromByte(buff[pointer])
				pointer++
				offset, err := strconv.ParseUint(bits, 2, 8)
				if err != nil {
					panic(err)
				}
				oldPointer = pointer
				pointer = int(offset) - 12
				break
			}
			if legthProcessed == labelLength {
				labels = append(labels, string(labelBytes))
				labelBytes = []byte{}
				break
			}
		}
		if buff[pointer] == byte(0) {
			if oldPointer != 0 {
				pointer = oldPointer
			} else {
				pointer++
			}
			record := binary.BigEndian.Uint16(buff[pointer : pointer+2])
			class := binary.BigEndian.Uint16(buff[pointer+2 : pointer+4])
			pointer += 4
			questions = append(questions, Question{
				Name:  strings.Join(labels, "."),
				Type:  recordType(record),
				Class: class,
			})
			labels = []string{}
		}
		if pointer >= len(buff) {
			break
		}
	}
	return questions
}

func (m *message) Bytes() []byte {
	header := m.header
	headerBytes := []byte{}
	// header := [12]byte{}
	// 16 bits => 2 bytes for id
	idBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes, header.id)
	// fmt.Println(idBytes, "IB")
	headerBytes = append(headerBytes, idBytes...)

	bits := ""

	if header.QR {
		bits += "1"
	} else {
		bits += "0"
	}

	// OPCODE
	bits += header.OPCODE

	if header.AA {
		bits += "1"
	} else {
		bits += "0"
	}

	// TC
	bits += "0"

	if header.RD {
		bits += "1"
	} else {
		bits += "0"
	}

	b, err := strconv.ParseUint(bits, 2, 8)
	if err != nil {
		panic(err)
	}

	headerBytes = append(headerBytes, byte(b))
	bits = ""

	if header.RA {
		bits += "1"
	} else {
		bits += "0"
	}

	// Z
	bits += "000"

	// RCODE
	if header.OPCODE == "0000" {
		bits += "0000"
	} else {
		bits += "0100"
	}

	b, err = strconv.ParseUint(bits, 2, 8)
	if err != nil {
		panic(err)
	}
	headerBytes = append(headerBytes, byte(b))

	queBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(queBytes, header.Questions)
	headerBytes = append(headerBytes, queBytes...)

	ansBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(ansBytes, header.Answers)
	headerBytes = append(headerBytes, ansBytes...)

	NSCountBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(NSCountBytes, header.NSCOUNT)
	headerBytes = append(headerBytes, NSCountBytes...)

	ARCOUNTSBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(ARCOUNTSBytes, header.ARCOunt)
	headerBytes = append(headerBytes, ARCOUNTSBytes...)

	msg := []byte{}
	msg = append(msg, headerBytes...)

	question := m.question
	for _, que := range question {
		QuestionBytes := []byte{}
		QuestionBytes = append(QuestionBytes, getLabelSequence(que.Name)...)
		QuestionBytes = append(QuestionBytes, []byte{0x0, byte(que.Type), 0x0, 0x1}...)
		msg = append(msg, QuestionBytes...)
	}

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
