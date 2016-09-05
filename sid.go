package evtxparser

import "fmt"

type Sid struct {
	Revision            uint8
	SubAuthorityCount   uint8
	IdentifierAuthority [6]uint8
	SubAuthority        []uint32
}

func (g Sid) String() string {
	s := fmt.Sprintf("S-%d", g.Revision)

	v := uint64(0)
	for _, ia := range g.IdentifierAuthority {
		v = v << 8
		v += uint64(ia)
	}
	s += fmt.Sprintf("-%d", v)

	for _, sa := range g.SubAuthority {
		s += fmt.Sprintf("-%d", sa)
	}

	return s
}
