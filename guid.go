package evtxparser

import (
	"encoding/binary"
	"fmt"
)

type Guid [16]byte

func (g Guid) String() string {
	return fmt.Sprintf("{%08X-%04X-%04X-%04X-%X}", binary.LittleEndian.Uint32(g[0:4]), binary.LittleEndian.Uint16(g[4:6]), binary.LittleEndian.Uint16(g[6:8]), g[8:10], g[10:])
}
