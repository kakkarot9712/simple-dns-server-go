package main

import (
	"encoding/binary"
	"fmt"
	"net"
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
	header   header
	question []question
	Answer   []answer
}

type header struct {
	Id        uint16
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

type question struct {
	Id     int
	Header header
	Name   string
	Type   recordType
	Class  uint16
}

type answer struct {
	Uid            int
	ResponseHeader header
	Name           string
	Type           recordType
	TTL            uint32
	Data           string
}

type DNSMessage struct {
	source    *net.UDPAddr
	questions []question
	id        uint16
	header    header
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

func ParseHeader(buff []byte) header {
	h := header{}
	if len(buff) < 12 {
		fmt.Println("Invalid Header passed!")
		return h
	}
	h.Id = binary.BigEndian.Uint16(buff[:2])

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

func ParseLabels(buff *[]byte, pointer int) ([]string, int) {
	labels := []string{}
	oldPointer := 0
	for {
		b := getOctetFromByte((*buff)[pointer])
		labelLength := int((*buff)[pointer])
		pointer++
		labelBytes := []byte{}
		// labels = append(labels, string(buff[pointer:pointer+labelLength]))
		legthProcessed := 0
		for {
			if b[:2] != "11" {
				labelBytes = append(labelBytes, (*buff)[pointer])
				pointer++
				legthProcessed++
			} else {
				bits := b[2:] + getOctetFromByte((*buff)[pointer])
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
				break
			}
		}
		if (*buff)[pointer] == byte(0) {
			if oldPointer != 0 {
				pointer = oldPointer
			} else {
				pointer++
			}
			return labels, pointer
		}
	}
}

func ParseAnswer(buff []byte, pointer int) ([]answer, int) {
	answers := []answer{}
	for {
		labels, newPointer := ParseLabels(&buff, pointer)
		pointer = newPointer
		record := binary.BigEndian.Uint16(buff[pointer : pointer+2])
		pointer += 2
		// Skip class
		// class := binary.BigEndian.Uint16(buff[pointer : pointer+2])
		pointer += 2
		TTL := binary.BigEndian.Uint32(buff[pointer : pointer+4])
		pointer += 4
		Length := binary.BigEndian.Uint16(buff[pointer : pointer+2])
		pointer += 2
		Data := buff[pointer : pointer+int(Length)]
		ipAddresss := []string{}
		for _, ip := range Data {
			numStr := fmt.Sprintf("%v", ip)
			ipAddresss = append(ipAddresss, numStr)
		}
		pointer += int(Length)
		answers = append(answers, answer{
			Name: strings.Join(labels, "."),
			Type: recordType(record),
			TTL:  TTL,
			Data: strings.Join(ipAddresss, "."),
		})
		if pointer >= len(buff) {
			return answers, pointer
		}
	}
}

func ParseQuestion(buff []byte, questionLenght int) ([]question, int) {
	questions := []question{}
	pointer := 0
	if questionLenght == 0 {
		return questions, pointer
	}
	for {
		labels, newPointer := ParseLabels(&buff, pointer)
		pointer = newPointer
		record := binary.BigEndian.Uint16(buff[pointer : pointer+2])
		class := binary.BigEndian.Uint16(buff[pointer+2 : pointer+4])
		pointer += 4
		questions = append(questions, question{
			Name:  strings.Join(labels, "."),
			Type:  recordType(record),
			Class: class,
		})
		if questionLenght == len(questions) {
			return questions, pointer
		}
	}
}

func (m *message) Bytes() []byte {
	header := m.header
	headerBytes := []byte{}
	// 16 bits => 2 bytes for id
	idBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes, header.Id)
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

	AnswerBytes := []byte{}

	answers := m.Answer

	for _, ans := range answers {
		AnswerBytes = append(AnswerBytes, getLabelSequence(ans.Name)...)
		recordBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(recordBytes, uint16(ans.Type))
		AnswerBytes = append(AnswerBytes, recordBytes...)

		classBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(classBytes, 1)
		AnswerBytes = append(AnswerBytes, classBytes...)

		TTLBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(TTLBytes, ans.TTL)
		AnswerBytes = append(AnswerBytes, TTLBytes...)

		RDataStr := strings.Split(ans.Data, ".")
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
		AnswerBytes = append(AnswerBytes, Length...)
		AnswerBytes = append(AnswerBytes, RData...)
	}

	msg = append(msg, AnswerBytes...)
	return msg
}
