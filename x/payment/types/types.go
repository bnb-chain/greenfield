package types

import "fmt"

func ParseReadPacket(readPacket string) (ReadPacket, error) {
	res, found := ReadPacket_value[readPacket]
	if !found {
		return ReadPacketFree, fmt.Errorf("invalid read packet: %s", readPacket)
	}
	return ReadPacket(res), nil
}
