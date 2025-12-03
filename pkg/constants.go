package websocket

const (
	NORMAL_CLOSURE             = 1000
	GOING_AWAY                 = 1001
	PROTOCOL_ERROR             = 1002
	UNSUPPORTED_DATA           = 1003
	RESERVED                   = 1004
	NO_STATUS_RCV              = 1005
	ABNORMAL_CLOSURE           = 1006
	INVALID_FRAME_PAYLOAD_DATA = 1007
	POLICY_VIOLATION           = 1008
	MESSAGE_TOO_BIG            = 1009
	MANDATORY_EXTENSION        = 1010
	INTERNAL_SERVER_ERROR      = 1011
	TLS_HANDSHAKE              = 1015
)

const (
	OP_CONTINUATION_FRAME = 0x0
	OP_TEXT_FRAME         = 0x1
	OP_BIANRY_FRAME       = 0x2
	OP_CLOSE_FRAME        = 0x8
	OP_PING_FRAME         = 0x9
	OP_PONG_Frame         = 0xA
)

const (
	MaxFrameSize   = 4096
	MaxMessageSize = 1024 * 1024
)
