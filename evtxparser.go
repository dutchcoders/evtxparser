package evtxparser

//go:generate stringer -type=Type

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"golang.org/x/text/encoding/unicode"
)

type Header struct {
	CurrentChunk uint64
	NextRecord   uint64
	Size1        uint32
	Size2        uint16

	Count uint16

	Minor uint16
	Major uint16

	Flags    uint32
	Checksum uint32
}

func (s *Header) Decode(d Decoder) {
	buff := [8]byte{}
	d.Copy(buff[:])

	Assert(bytes.Equal(buff[0:7], []byte("ElfFile")))

	d.Skip(8)
	s.CurrentChunk = d.Uint64()
	s.NextRecord = d.Uint64()
	s.Size1 = d.Uint32()

	s.Minor = d.Uint16()
	s.Major = d.Uint16()

	s.Size2 = d.Uint16()
	s.Count = d.Uint16()

	d.Skip(76)

	s.Flags = d.Uint32()
	s.Checksum = d.Uint32()

	d.Skip(3968)
}

type TemplateDefinition struct {
	Pointer uint32
	Guid    [16]byte

	Length uint32
	Flag   uint8

	ElementNode *ElementNode
}

func (s *TemplateDefinition) Dump(sa SubstitutionArray) {
	s.ElementNode.Dump(sa)
}

func (s *TemplateDefinition) Decode(d Decoder) {
	s.Pointer = d.Uint32()

	d.Copy(s.Guid[:])

	s.Length = d.Uint32()
	s.Flag = d.Uint8()

	d.Skip(3)

	en := &ElementNode{}
	en.Decode(d)
	s.ElementNode = en

	Assert(d.Uint8() == 0x00)
}

type Chunk struct {
}

func (ch *Chunk) Decode(d Decoder) {
	hdr := ChunkHeader{}
	hdr.Decode(d)

	// string table
	d.Skip(256)

	// template table
	d.Skip(128)

	for i := 0; i < int(hdr.LastRecord-hdr.FirstRecord); i++ {
		b := AuditRecord{}
		b.Decode(d)
	}
}

type ChunkHeader struct {
	Magic         [8]byte
	Size          uint32
	FirstRecord   uint64
	LastRecord    uint64
	FirstRecordID uint64
	LastRecordID  uint64
	Flags         uint32
	Checksum      uint32
	PtrToLast     uint32
	PtrToNext     uint32
}

var (
	MagicElfChunk    = []byte{0x45, 0x6c, 0x66, 0x43, 0x68, 0x6e, 0x6b, 0x0}
	MagicAuditRecord = []byte{0x2a, 0x2a, 0x0, 0x0}
)

func (ch *ChunkHeader) Decode(d Decoder) {
	d.Copy(ch.Magic[:])

	Assert(bytes.Equal(ch.Magic[:], MagicElfChunk))

	ch.FirstRecord = d.Uint64()
	ch.LastRecord = d.Uint64()
	ch.FirstRecordID = d.Uint64()
	ch.LastRecordID = d.Uint64()

	ch.Size = d.Uint32()

	ch.PtrToLast = d.Uint32()
	ch.PtrToNext = d.Uint32()

	d.Uint32()

	d.Skip(64)

	ch.Flags = d.Uint32()
	ch.Checksum = d.Uint32()
}

type AuditRecord struct {
	Length   uint32
	RecordID uint64
	Time     time.Time
	Magic    [4]byte
}

func (ar *AuditRecord) Decode(d Decoder) {
	start := d.Offset()
	_ = start

	d.Copy(ar.Magic[:])

	Assert(bytes.Equal(ar.Magic[:], MagicAuditRecord))

	ar.Length = d.Uint32()

	ar.RecordID = d.Uint64()

	lowDateTime := int64(d.Uint32())
	highDateTime := int64(d.Uint32())
	nsec := int64(highDateTime)<<32 + int64(lowDateTime)
	nsec -= 116444736000000000
	nsec *= 100
	ar.Time = time.Unix(0, nsec)

	s := Stream{}
	s.Decode(d)

	s.Dump()

	d.Skip(int(ar.Length) - (d.Offset() - start))
}

func Decode(Decoder) {
}

type ElementNode struct {
	Length uint32

	StringStructure *StringStructure
	Attributes      *Attributes
	Children        *Children
}

func (s *ElementNode) Decode(d Decoder) {
	Assert(d.PeekUint8() == 0x1 || d.PeekUint8() == 0x41)

	d.Uint8()
	d.Uint16()

	s.Length = d.Uint32()

	stringPtr := d.Uint32()
	if (int(stringPtr)) == d.Offset() {
		ss := &StringStructure{
			Ptr: stringPtr,
		}
		ss.Decode(d)
		s.StringStructure = ss
	} else {
		s.StringStructure = pointers[stringPtr].(*StringStructure)
	}

	if d.PeekUint8() == 0x02 {
	} else if d.PeekUint8() == 0x03 {
	} else {
		aa := &Attributes{}
		aa.Decode(d)
		s.Attributes = aa
	}

	switch d.PeekUint8() {
	case 0x2:
		Assert(d.Uint8() == 0x2)

		// children
		aa := &Children{}
		aa.Decode(d)
		s.Children = aa
	case 0x3:
		Assert(d.Uint8() == 0x3)

	default:
		d.Dump()
		panic("Unexpected")
		// attributes
	}
}

func (s *ElementNode) Dump(sa SubstitutionArray) {
	if s.StringStructure != nil {
		fmt.Printf("<%s", s.StringStructure.String())
	}

	if s.Attributes != nil {
		s.Attributes.Dump(sa)
	}

	fmt.Printf(">")

	if s.Children != nil {
		for _, child := range *s.Children {
			_ = child
			if en, ok := child.(*ElementNode); ok {
				fmt.Printf("\n")
				en.Dump(sa)
			}

			if v, ok := child.(*Value); ok {
				fmt.Printf("%s", v.String())
			}

			if v, ok := child.(*Substitution); ok {
				fmt.Printf("%s", v.Dump(sa))
			}
		}
	}

	if s.StringStructure != nil {
		fmt.Printf("</%s>\n", s.StringStructure.String())
	}
}

type Children []interface{}

func (s *Children) Decode(d Decoder) {
	for {
		if d.PeekUint8() == 4 {
			d.Uint8()
			return
		}

		if d.PeekUint8() == 0x5 {
			v := &Value{}
			v.Decode(d)
			*s = append(*s, v)
		} else if d.PeekUint8() == 0x0d || d.PeekUint8() == 0x0e {
			v := &Substitution{}
			v.Decode(d)
			*s = append(*s, v)
		} else if d.PeekUint8() == 0x41 || d.PeekUint8() == 0x01 {
			en := &ElementNode{}
			en.Decode(d)
			*s = append(*s, en)
		} else {
			panic("Unexpected")
		}
	}
}

type Attributes []*Attribute

type Attribute struct {
	StringStructure *StringStructure

	Value        *Value
	Substitution *Substitution
}

func (s *Attributes) Dump(sa SubstitutionArray) {
	for _, attribute := range *s {
		fmt.Printf(" ")

		if attribute.StringStructure != nil {
			fmt.Printf("%s=", attribute.StringStructure.String())
		}

		if attribute.Value != nil {
			fmt.Printf("\"%s\"", attribute.Value.String())
		} else if attribute.Substitution != nil {
			fmt.Printf("\"%s\"", attribute.Substitution.Dump(sa))
		}
	}
}

func (s *Attributes) Decode(d Decoder) {
	_ = d.Uint32() // length

	for {
		flag := d.Uint8()

		attribute := &Attribute{}

		stringPtr := d.Uint32()
		if (int(stringPtr)) == d.Offset() {
			ss := &StringStructure{
				Ptr: stringPtr,
			}
			ss.Decode(d)

			attribute.StringStructure = ss
		} else {
			attribute.StringStructure = pointers[stringPtr].(*StringStructure)
		}

		switch d.PeekUint8() {
		case 0x5:
			v := &Value{}
			v.Decode(d)
			attribute.Value = v
		case 0xe:
			fallthrough
		case 0xd:
			v := &Substitution{}
			v.Decode(d)
			attribute.Substitution = v
		}

		*s = append(*s, attribute)

		if flag == 0x06 {
			return
		}
	}
}

type Substitution struct {
	Index uint16
	Type  Type
}

func (s *Substitution) Decode(d Decoder) {
	d.Uint8() // 0xe or 0xd

	s.Index = d.Uint16()
	s.Type = Type(d.Uint8())
}

func (s *Substitution) Dump(sa SubstitutionArray) string {
	if v, ok := sa[uint32(s.Index)]; !ok {
		return "Unknown"
	} else {
		switch v := v.(type) {
		case nil:
			return ""
		case bool:
			if v {
				return "true"
			} else {
				return "false"
			}
		case uint64:
			return fmt.Sprintf("%d", v)
		case int64:
			return fmt.Sprintf("%d", v)
		case uint32:
			return fmt.Sprintf("%d", v)
		case int32:
			return fmt.Sprintf("%d", v)
		case uint16:
			return fmt.Sprintf("%d", v)
		case int16:
			return fmt.Sprintf("%d", v)
		case uint8:
			return fmt.Sprintf("%d", v)
		case int8:
			return fmt.Sprintf("%d", v)
		case string:
			return fmt.Sprintf("%s", v)
		case Stream:
			v.Dump()
		case Sid:
			return fmt.Sprintf("%s", v.String())
		case Guid:
			return fmt.Sprintf("%s", v.String())
		case time.Time:
			return fmt.Sprintf("%s", v.Format(time.RFC3339))
		default:
			return fmt.Sprintf("%#v", v)
		}

	}

	return ""
}

func (s *Substitution) String() string {
	return fmt.Sprintf("Substitution: %d %d", s.Index, s.Type)
}

type Stream struct {
	TemplateID uint32
	Ptr        uint32

	TemplateDefinition *TemplateDefinition
	SubstitutionArray  SubstitutionArray
}

func (s *Stream) Dump() {
	if s.TemplateDefinition != nil {
		s.TemplateDefinition.Dump(s.SubstitutionArray)
	}
}

func (s *Stream) Decode(d Decoder) {
	if d.PeekUint8() == 0xf {
		Assert(d.Uint8() == 0xf)
		Assert(d.Uint8() == 0x1)
		Assert(d.Uint8() == 0x1)
		Assert(d.Uint8() == 0x0)
	}

	Assert(d.Uint8() == 0xc)
	Assert(d.Uint8() == 0x1)

	// https://static1.squarespace.com/static/510d93d8e4b060f86e6fdf2d/t/5328923ee4b0bea727f8aa9b/1395167806310/Windows+7+Audit+Format+v10.pdf
	s.TemplateID = d.Uint32()

	s.Ptr = d.Uint32()
	if (int(s.Ptr)) == d.Offset() {
		t := &TemplateDefinition{}
		t.Decode(d)

		s.TemplateDefinition = t
	} else {
		// s.TemplateDefinition =
	}

	t := SubstitutionArray{}
	t.Decode(d)
	s.SubstitutionArray = t
}

type SubstitutionArray map[uint32]interface{}

func (s SubstitutionArray) Decode(d Decoder) {
	count := d.Uint32()

	iis := make([]IndexInfo, count)
	for i := uint32(0); i < count; i++ {
		ii := IndexInfo{}
		ii.Decode(d)

		iis[i] = ii
	}

	for i := uint32(0); i < count; i++ {
		length := iis[i].Length

		switch iis[i].Type {
		case EvtVarTypeNull:
			s[i] = nil
			continue
		case EvtVarTypeString:
			buff6 := make([]byte, length)
			d.Copy(buff6[:])

			utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
			if val, err := utf16.Bytes(buff6[:]); err == nil {
				s[i] = string(val)
			} else {
			}
			continue
		case EvtVarTypeAnsiString:
			// fmt.Println("EvtVarTypeAnsiString")
		case EvtVarTypeSByte:
			s[i] = d.Uint8()
			continue
		case EvtVarTypeByte:
			s[i] = d.Int8()
			continue
		case EvtVarTypeInt16:
			s[i] = d.Int16()
			continue
		case EvtVarTypeUInt16:
			s[i] = d.Uint16()
			continue
		case EvtVarTypeInt32:
			s[i] = d.Int32()
			continue
		case EvtVarTypeUInt32:
			s[i] = d.Uint32()
			continue
		case EvtVarTypeInt64:
			s[i] = d.Int64()
			continue
		case EvtVarTypeUInt64:
			s[i] = d.Uint64()
			continue
		case EvtVarTypeSingle:
			// fmt.Println("EvtVarTypeSingle")
		case EvtVarTypeDouble:
			// fmt.Println("EvtVarTypeDouble")
		case EvtVarTypeBoolean:
			// fmt.Println("EvtVarTypeBoolean")
		case EvtVarTypeBinary:
			// fmt.Println("EvtVarTypeBinary")
		case EvtVarTypeGuid:
			guid := Guid{}
			d.Copy(guid[:])
			s[i] = guid
			continue
		case EvtVarTypeSizeT:
			// fmt.Println("EvtVarTypeSizeT")
		case EvtVarTypeFileTime:
			lowDateTime := int64(d.Uint32())
			highDateTime := int64(d.Uint32())
			nsec := int64(highDateTime)<<32 + int64(lowDateTime)
			nsec -= 116444736000000000
			nsec *= 100
			s[i] = time.Unix(0, nsec)
			continue
		case EvtVarTypeSysTime:
			lowDateTime := int64(d.Uint32())
			highDateTime := int64(d.Uint32())
			nsec := int64(highDateTime)<<32 + int64(lowDateTime)
			nsec -= 116444736000000000
			nsec *= 100
			s[i] = time.Unix(0, nsec)
		case EvtVarTypeSid:
			sid := Sid{}
			sid.Revision = d.Uint8()
			sid.SubAuthorityCount = d.Uint8()
			d.Copy(sid.IdentifierAuthority[:])

			sid.SubAuthority = make([]uint32, sid.SubAuthorityCount)
			for i := uint8(0); i < sid.SubAuthorityCount; i++ {
				sid.SubAuthority[i] = d.Uint32()
			}

			s[i] = sid
			continue
		case EvtVarTypeHexInt32:
			s[i] = fmt.Sprintf("0x%X", d.Uint32())
			continue
		case EvtVarTypeHexInt64:
			s[i] = fmt.Sprintf("0x%X", d.Uint64())
			continue
		case EvtVarTypeEvtHandle:
			data := make([]byte, iis[i].Length)
			d.Copy(data)

			s[i] = fmt.Sprintf("EvtVarTypeEvtHandle  %x", data)
			continue
		case BinaryXmlStream:
			startOffset := d.Offset()

			stream := Stream{}
			stream.Decode(d)

			d.Seek(startOffset + int(iis[i].Length))

			s[i] = stream
			continue
		case EvtVarTypeEvtXml:
			s[i] = "EvtVarTypeEvtXML"
		}

		/*
			if d.Offset() != startOffset+int(iis[i].Length) {
				panic("Not equal")
			}
		*/

		d.Skip(int(iis[i].Length))
	}
}

type IndexInfo struct {
	Length uint16
	Type   Type
}

func (s *IndexInfo) Decode(d Decoder) {
	s.Length = d.Uint16()
	s.Type = Type(d.Uint8())

	d.Uint8()
}

type Type uint8

const (
	EvtVarTypeNull       Type = 0x0
	EvtVarTypeString          = 0x1
	EvtVarTypeAnsiString      = 0x2
	EvtVarTypeSByte           = 0x3
	EvtVarTypeByte            = 0x4
	EvtVarTypeInt16           = 0x5
	EvtVarTypeUInt16          = 0x6
	EvtVarTypeInt32           = 0x7
	EvtVarTypeUInt32          = 0x8
	EvtVarTypeInt64           = 0x9
	EvtVarTypeUInt64          = 0xa
	EvtVarTypeSingle          = 0xb
	EvtVarTypeDouble          = 0xc
	EvtVarTypeBoolean         = 0xd
	EvtVarTypeBinary          = 0xe
	EvtVarTypeGuid            = 0xf
	EvtVarTypeSizeT           = 0x10
	EvtVarTypeFileTime        = 0x11
	EvtVarTypeSysTime         = 0x12
	EvtVarTypeSid             = 0x13
	EvtVarTypeHexInt32        = 0x14
	EvtVarTypeHexInt64        = 0x15
	EvtVarTypeEvtHandle       = 0x20
	BinaryXmlStream           = 0x21
	EvtVarTypeEvtXml          = 0x23
)

func (t Type) String() string {
	switch t {
	case EvtVarTypeNull:
		return "EvtVarTypeNull"
	case EvtVarTypeString:
		return "EvtVarTypeString"
	case EvtVarTypeAnsiString:
		return "EvtVarTypeAnsiString"
	case EvtVarTypeSByte:
		return "EvtVarTypeSByte"
	case EvtVarTypeByte:
		return "EvtVarTypeByte"
	case EvtVarTypeInt16:
		return "EvtVarTypeInt16"
	case EvtVarTypeUInt16:
		return "EvtVarTypeUInt16"
	case EvtVarTypeInt32:
		return "EvtVarTypeInt32"
	case EvtVarTypeUInt32:
		return "EvtVarTypeUInt32"
	case EvtVarTypeInt64:
		return "EvtVarTypeInt64"
	case EvtVarTypeUInt64:
		return "EvtVarTypeUInt64"
	case EvtVarTypeSingle:
		return "EvtVarTypeSingle"
	case EvtVarTypeDouble:
		return "EvtVarTypeDouble"
	case EvtVarTypeBoolean:
		return "EvtVarTypeBoolean"
	case EvtVarTypeBinary:
		return "EvtVarTypeBinary"
	case EvtVarTypeGuid:
		return "EvtVarTypeGuid"
	case EvtVarTypeSizeT:
		return "EvtVarTypeSizeT"
	case EvtVarTypeFileTime:
		return "EvtVarTypeFileTime"
	case EvtVarTypeSysTime:
		return "EvtVarTypeSysTime"
	case EvtVarTypeSid:
		return "EvtVarTypeSid"
	case EvtVarTypeHexInt32:
		return "EvtVarTypeHexInt32"
	case EvtVarTypeHexInt64:
		return "EvtVarTypeHexInt64"
	case EvtVarTypeEvtHandle:
		return "EvtVarTypeEvtHandle"
	case BinaryXmlStream:
		return "BinaryXmlStream"
	case EvtVarTypeEvtXml:
		return "EvtVarTypeEvtXml"
	}

	return "Unknown format"
}

type Value struct {
	Type   uint8
	Length uint16
	Data   []byte
}

func (s *Value) Decode(d Decoder) {
	Assert(d.Uint8() == 0x5)
	s.Type = d.Uint8()
	s.Length = d.Uint16()

	s.Data = make([]byte, s.Length*2)
	d.Copy(s.Data[:])
}

func (s *Value) String() string {
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	if val, err := utf16.Bytes(s.Data); err == nil {
		return string(val)
	} else {
		return ""
	}
}

var pointers map[uint32]interface{} = map[uint32]interface{}{}

type StringStructure struct {
	Ptr      uint32
	NextPtr  uint32
	Checksum uint16
	Count    uint16
	Data     []byte
}

func (s *StringStructure) String() string {
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	if val, err := utf16.Bytes(s.Data[:len(s.Data)-2]); err == nil {
		return string(val)
	} else {
		return ""
	}
}

func (s *StringStructure) Decode(d Decoder) {
	s.NextPtr = d.Uint32()
	s.Checksum = d.Uint16()

	s.Count = d.Uint16()

	s.Data = make([]byte, s.Count*2+2)
	d.Copy(s.Data)

	pointers[s.Ptr] = s
}

func Assert(b bool) {
	if b {
		return
	}

	panic("expr")
}

type parser struct {
}

func Parse(data []byte) *parser {
	p := parser{}

	d := NewDefaultDecoder(data, binary.LittleEndian)

	h := Header{}
	h.Decode(d)

	for i := 0; i < int(h.Count); i++ {
		subDecoder := d.NewDecoder()

		ch := Chunk{}
		ch.Decode(subDecoder)

		d.Skip(0x10000)
	}

	return &p
}
