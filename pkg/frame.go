package websocket

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
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

type Message struct {
	binary          bool // true = binary format false = text
	ApplicationData []byte
}

func ParseNetworkFrame(conn *WsConnection) (*Frame, []error) {
	frame := &Frame{}
	var errors []error

	header := make([]byte, 2)
	if _, err := io.ReadFull(conn.Conn, header); err != nil {
		return nil, append(errors, err)
	}
	// BYTE 0
	start, end := 0, 1
	p := header[start:end]

	frame.fin = (p[0] & 0x80) != 0
	frame.rsv1 = (p[0] & 0x40) != 0
	frame.rsv2 = (p[0] & 0x20) != 0
	frame.rsv3 = (p[0] & 0x10) != 0
	frame.Opcode = p[0] & 0x0F

	// BYTE 1
	start = end
	end = start + 1
	p = header[start:end]
	frame.Mask = (p[0] & 0x80) != 0
	payloadLen7 := p[0] & 0x7F

	if (frame.Opcode >= 0x3 && frame.Opcode <= 0x7) || (frame.Opcode >= 0xB && frame.Opcode <= 0xF) {
		return nil, append(errors, fmt.Errorf("reserved opcode: 0x%X", frame.Opcode))
	}

	// PAYLOAD LENGTH
	if payloadLen7 < 126 {
		frame.payloadLen = int64(payloadLen7)
		start = end
		end = start
	} else if payloadLen7 == 126 {
		lenBuf := make([]byte, 2)
		if _, err := io.ReadFull(conn.Conn, lenBuf); err != nil {
			return nil, append(errors, err)
		}
		frame.payloadLen = int64(binary.BigEndian.Uint16(lenBuf))
	} else {
		lenBuf := make([]byte, 8)
		if _, err := io.ReadFull(conn.Conn, lenBuf); err != nil {
			return nil, append(errors, err)
		}
		frame.payloadLen = int64(binary.BigEndian.Uint64(lenBuf))
	}

	if frame.Opcode >= 0x8 { // Control frame
		if frame.payloadLen > 125 { // Changed from payloadLen7
			return nil, append(errors, fmt.Errorf("control frame payload too large: %d", frame.payloadLen))
		}
		if !frame.fin {
			return nil, append(errors, fmt.Errorf("control frame must not be fragmented"))
		}
	}

	// MASK KEY
	if frame.Mask {
		maskBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn.Conn, maskBuf); err != nil {
			return nil, append(errors, err)
		}
		frame.MaskKey = maskBuf
	}

	// PAYLOAD
	payloadBuf := make([]byte, frame.payloadLen)
	if _, err := io.ReadFull(conn.Conn, payloadBuf); err != nil {
		return nil, append(errors, err)
	}

	// Unmasking
	frame.PayloadData = payloadBuf
	if frame.Mask {
		for i := range frame.PayloadData {
			frame.PayloadData[i] ^= frame.MaskKey[i%4]
		}
	}

	return frame, errors
}

func (f *Frame) ComposeNetworkFrame() []byte {
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
	} else if f.payloadLen <= 0xFFFF {
		b1 |= 126
		frame = append(frame, b1)
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(f.payloadLen))
		frame = append(frame, buf...)
	} else {
		b1 |= 127
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
		fin:         true,
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
		fin:         true,
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
		fin:         true,
		Opcode:      0xA, // Opcode for Pong Frame
		Mask:        false,
		payloadLen:  int64(len(data)),
		PayloadData: data,
	}
}

func ParseMessage(conn *WsConnection) *Message {
	var message *Message
	for {
		frame, errs := ParseNetworkFrame(conn)
		if errs != nil {
			for _, v := range errs {
				log.Println(v.Error())
			}
			closeFrame := NewCloseFrame("Malfunctioned frames deteched closing the connection")
			conn.Conn.Write(closeFrame.ComposeNetworkFrame())
			conn.Conn.Close()
			return nil
		}
		if frame.fin == true {
			switch frame.Opcode {
			case 0x8:
				closeFrame := NewCloseFrame("Closing Frame Detected: closign the connection")
				conn.Conn.Write(closeFrame.ComposeNetworkFrame())
				conn.Conn.Close()
				return nil
			case 0x9:
				pongFrame := NewPongFrame("replying to ping frame")
				conn.Conn.Write(pongFrame.ComposeNetworkFrame())
				continue
			case 0xA:
				// will be implemented later for now just log and continue
				fmt.Println("will be implemented")
				continue
			default:
				if message == nil {
					if frame.Opcode == 0x0 {
						closeFrame := NewCloseFrame("Orphan Continuation Frame")
						conn.Conn.Write(closeFrame.ComposeNetworkFrame())
						conn.Conn.Close()
						return nil
					}
					message = &Message{}
					if frame.Opcode == 0x2 {
						message.binary = true
					}
				}
				if !checkAndAppendPayload(message, frame, conn) {
					return nil
				}
				return message
			}
		} else {
			switch frame.Opcode {
			case 0x1:
				if message != nil {
					closeFrame := NewCloseFrame("Cannot start a new message while processing a fragmented message")
					conn.Conn.Write(closeFrame.ComposeNetworkFrame())
					conn.Conn.Close()
					return nil
				} else {
					message = &Message{}
					message.binary = false
					if !checkAndAppendPayload(message, frame, conn) {
						return nil
					}
				}
			case 0x8:
				closeFrame := NewCloseFrame("Closing Frame Detected: closign the connection")
				conn.Conn.Write(closeFrame.ComposeNetworkFrame())
				conn.Conn.Close()
				return nil
			case 0x9:
				pongFrame := NewPongFrame("replying to ping frame")
				conn.Conn.Write(pongFrame.ComposeNetworkFrame())
				continue
			case 0xA:
				// will be implemented later for now just log and continue
				fmt.Println("will be implemented")
				continue
			case 0x2:
				if message != nil {
					closeFrame := NewCloseFrame("Cannot start a new message while processing a fragmented message")
					conn.Conn.Write(closeFrame.ComposeNetworkFrame())
					conn.Conn.Close()
					return nil
				} else {
					message = &Message{}
					message.binary = true
					if !checkAndAppendPayload(message, frame, conn) {
						return nil
					}
				}
			case 0x0:
				if message == nil {
					closeFrame := NewCloseFrame("Orphan Continuation Frame")
					conn.Conn.Write(closeFrame.ComposeNetworkFrame())
					conn.Conn.Close()
					return nil
				}
				if !checkAndAppendPayload(message, frame, conn) {
					return nil
				}
			}
		}
	}
}

func checkAndAppendPayload(message *Message, frame *Frame, conn *WsConnection) bool {
	if int64(len(message.ApplicationData))+int64(len(frame.PayloadData)) > int64(MaxMessageSize) {
		log.Printf("maximum size of the message exceeded from %v", conn.Conn.RemoteAddr())
		closeFrame := NewCloseFrame(fmt.Sprintf("%d Maximum message size exceeded\n", MESSAGE_TOO_BIG))
		conn.Conn.Write(closeFrame.ComposeNetworkFrame())
		conn.Conn.Close()
		return false
	}
	message.ApplicationData = append(message.ApplicationData, frame.PayloadData...)
	return true
}
