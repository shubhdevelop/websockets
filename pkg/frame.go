package websocket

import (
	"encoding/binary"
)

type Frame struct {
	fin    bool
	rsv1   bool
	rsv2   bool
	rsv3   bool
	Opcode byte
	Mask   bool
	/*
	 if 0-125: that is the payload length
	 if 126: following 2 bytes interpreted as a 16 bit unsigned
	 if 127: following 8 bytes as 64 bit unsigned integer
	*/
	payloadLen int64
	MaskKey    []byte
	/*
	 * payload_data - Extension data = pplication data
	 *
	 */
	PayloadData []byte
}

func ParseNetworkFrame(conn *WsConnection) (*Frame, []error) {
	frame := &Frame{}
	var errors []error

	b := make([]byte, 4096)
	n, err := conn.Conn.Read(b)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}
	b = b[:n]

	// BYTE 0
	start, end := 0, 1
	p := b[start:end]

	frame.fin = (p[0] & 0x80) != 0
	frame.rsv1 = (p[0] & 0x40) != 0
	frame.rsv2 = (p[0] & 0x20) != 0
	frame.rsv3 = (p[0] & 0x10) != 0
	frame.Opcode = p[0] & 0x0F

	// BYTE 1
	start = end
	end = start + 1
	p = b[start:end]

	frame.Mask = (p[0] & 0x80) != 0
	payloadLen7 := p[0] & 0x7F

	// PAYLOAD LENGTH
	if payloadLen7 < 126 {
		frame.payloadLen = int64(payloadLen7)
		start = end // move past second byte
		end = start // no extra bytes
	} else if payloadLen7 == 126 {
		start = end
		end = start + 2
		frame.payloadLen = int64(binary.BigEndian.Uint16(b[start:end]))
	} else { // payloadLen7 == 127
		start = end
		end = start + 8
		frame.payloadLen = int64(binary.BigEndian.Uint64(b[start:end]))
	}

	// MASK KEY
	if frame.Mask {
		start = end
		end = start + 4
		frame.MaskKey = b[start:end]
	}

	// PAYLOAD
	start = end
	end = start + int(frame.payloadLen)

	frame.PayloadData = make([]byte, frame.payloadLen)
	copy(frame.PayloadData, b[start:end])

	if frame.Mask {
		for i := range frame.PayloadData {
			frame.PayloadData[i] ^= frame.MaskKey[i%4]
		}
	}

	return frame, errors
}

func (f *Frame) ComponseNetworkFrame() []byte {
	frame := make([]byte, 0, 14)
	b0 := byte(0)
	if f.fin {
		b0 |= 0x80
	}
	if f.rsv1 {
		b0 |= 0x40
	}
	if f.rsv2 {
		b0 |= 0x20
	}
	if f.rsv3 {
		b0 |= 0x10
	}
	b0 |= (f.Opcode & 0x0F)
	frame = append(frame, b0)
	b1 := byte(0)
	if f.Mask {
		b1 |= 0x80
	}
	if f.payloadLen <= 125 {
		b1 |= byte(f.payloadLen)
		frame = append(frame, b1)
	} else if f.payloadLen <= 0xFFF {
		b1 |= 126
		frame = append(frame, b1)
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(f.payloadLen))
		frame = append(frame, buf...)
	} else {
		b1 |= 126
		frame = append(frame, b1)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(f.payloadLen))
		frame = append(frame, buf...)
	}
	if f.Mask {
		frame = append(frame, f.MaskKey[:]...)
	}
	if f.Mask {
		for i, v := range f.PayloadData {
			frame = append(frame, v^f.MaskKey[i%4])
		}
	} else {
		frame = append(frame, f.PayloadData...)
	}
	return frame
}

func NewFrame(message any) *Frame {
	var data []byte
	var Opcode byte

	switch v := message.(type) {
	case string:
		data = []byte(v)
		Opcode = 0x1
	case []byte:
		data = v
		Opcode = 0x1
	default:
		panic("unsupported message type")
	}

	return &Frame{
		fin:         true,
		Opcode:      Opcode,
		Mask:        false,
		payloadLen:  int64(len(data)),
		PayloadData: data,
	}
}
func NewCloseFrame(body string) *Frame {
	var data []byte

	data = []byte(body)

	return &Frame{
		fin:         false,
		Opcode:      0x8, // Opcode for Close Frame
		Mask:        false,
		payloadLen:  int64(len(data)),
		PayloadData: data,
	}
}
func NewPingFrame(body string) *Frame {
	var data []byte

	data = []byte(body)

	return &Frame{
		fin:         false,
		Opcode:      0x9, // Opcode for Ping Frame
		Mask:        false,
		payloadLen:  int64(len(data)),
		PayloadData: data,
	}
}
func NewPongFrame(body string) *Frame {
	var data []byte

	data = []byte(body)

	return &Frame{
		fin:         false,
		Opcode:      0xA, // Opcode for Pong Frame
		Mask:        false,
		payloadLen:  int64(len(data)),
		PayloadData: data,
	}
}
