package comms

// Packet - Describes a minimum reach IPC packet
type Packet struct {
	Header Header
	Data   []byte
}

// Header - Description of each IPC packet
type Header struct {
	MsgId  uint32
	Target string
	Action string
	Ack    bool
}

// GpioData - Describes a
type GpioData struct {
	Addr uint32
	Val  uint32
}

// BuildPacket returns a packet with appropriate command and data assigned.
func BuildPacket(target string, action string, data []byte) Packet {
	packet := basicPacket(target, action)
	packet.Data = data
	return packet
}

func basicPacket(target string, action string) Packet {
	var packet Packet
	packet.Header.Target = target
	packet.Header.Action = action
	return packet
}

// BuildResponsePacket returns a packet with appropriate command and data assigned.
func BuildResponsePacket(requestHeader Header, data []byte) Packet {
	response := Packet{Header: requestHeader, Data: data}
	response.Header.Ack = true
	return response
}
