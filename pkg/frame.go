package websocket

import "encoding/binary"

type Frame struct {
	fin     bool
	rsv1    bool
	rsv2    bool
	rsv3    bool
	opcode  byte
	mask    bool
	maskKey [4]byte
	/*
	 if 0-125: that is the payload length
	 if 126: following 2 bytes interpreted as a 16 bit unsigned
	 if 127: following 8 bytes as 64 bit unsigned integer
	*/
	payloadLen int64
	/*
	 * payload_data - Extension data = pplication data
	 *
	 */
	payloadData []byte
}

func ParseNetworkFrame([]byte) *Frame {
	frame := &Frame{}
	return frame
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
	b0 |= (f.opcode & 0x0F)
	frame = append(frame, b0)
	b1 := byte(0)
	if f.mask {
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
	if f.mask {
		frame = append(frame, f.maskKey[:]...)
	}
	if f.mask {
		for i, v := range f.payloadData {
			frame = append(frame, v^f.maskKey[i%4])
		}
	} else {
		frame = append(frame, f.payloadData...)
	}
	return frame
}

func NewFrame(message any) *Frame {
	var data []byte
	var opcode byte

	switch v := message.(type) {
	case string:
		data = []byte(v)
		opcode = 0x1
	case []byte:
		data = v
		opcode = 0x1
	default:
		panic("unsupported message type")
	}

	return &Frame{
		fin:         true,
		opcode:      opcode,
		mask:        false,
		payloadLen:  int64(len(data)),
		payloadData: data,
	}
}
