package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/text/encoding/unicode"
)

type ElfFile struct {
}

type EvtChunk struct {
}

type TemplateDefinition struct {
}

type AuditRecord struct {
}

type ElementNode struct {
}

func Decode(Decoder) {
}

func (s *ElementNode) Decode(d Decoder) {
	fmt.Println("Elementnode")

	Assert(d.PeekUint8() == 0x1 || d.PeekUint8() == 0x41)
	fmt.Println("unknown", d.Uint8())

	fmt.Println("unknown", d.Uint16())
	fmt.Println("length", d.Uint32())

	d.Dump()
	fmt.Println("string ptr", d.Uint32())

	// string structure
	ss := &StringStructure{}
	ss.Decode(d)

	d.Dump()

	if d.PeekUint8() == 0x02 {
	} else if d.PeekUint8() == 0x03 {
	} else {
		fmt.Println("Elementnode: attributes")
		d.Dump()
		aa := &Attributes{}
		aa.Decode(d)
		d.Dump()
	}

	switch d.PeekUint8() {
	case 0x2:
		Assert(d.Uint8() == 0x2)

		// children
		aa := &Children{}
		aa.Decode(d)
	case 0x3:
		Assert(d.Uint8() == 0x3)

	default:
		// attributes
	}
	fmt.Println("edon tnetmelE")

}

type Children struct {
}

func (s *Children) Decode(d Decoder) {
	fmt.Println("Children")
	defer fmt.Println("nerdlihC")
	for {
		fmt.Println("Children: Peek uint*")
		d.Dump()
		if d.PeekUint8() == 4 {
			fmt.Println("Children: exit")
			d.Uint8()
			return
		}

		if d.PeekUint8() == 0x5 {
			fmt.Println("Children: Value")
			// value
			v := Value{}
			v.Decode(d)
		} else if d.PeekUint8() == 0x0d || d.PeekUint8() == 0x0e {
			fmt.Println("Children: Subtitution")
			v := Substitution{}
			v.Decode(d)

			// substituion
		} else if d.PeekUint8() == 0x41 || d.PeekUint8() == 0x01 {
			fmt.Println("Children: Element Node")
			// element node
			en := &ElementNode{}
			en.Decode(d)
		} else {
			panic("Unexpected")
		}
	}
}

type Attributes struct {
}

func (s *Attributes) Decode(d Decoder) {
	fmt.Println("Attributes")
	fmt.Println("length", d.Uint32())
	fmt.Println("flag", d.Uint8())

	fmt.Println("string ptr", d.Uint32())

	d.Dump()

	//		fmt.Println("value", d.Uint8())
	ss := &StringStructure{}
	ss.Decode(d)

	switch d.PeekUint8() {
	case 0x5:
		v := Value{}
		v.Decode(d)
	case 0xe:
		fallthrough
	case 0xd:
		v := Substitution{}
		v.Decode(d)
	}

	fmt.Println("setubirttA")
}

type Substitution struct {
}

func (s *Substitution) Decode(d Decoder) {
	fmt.Println("Substitution")

	d.Uint8() // 0xe or 0xd

	index := d.Uint16()
	fmt.Println("index", index)

	// this is in spec
	/*
		type_ := d.Uint8()
		fmt.Println("type_", type_)
	*/

	fmt.Println("noitutitsbuS")
}

type Value struct {
}

func (s *Value) Decode(d Decoder) {
	fmt.Println("Value")

	Assert(d.Uint8() == 0x5)
	fmt.Println("Type", d.Uint8())
	length := d.Uint16()

	buff6 := make([]byte, length*2)
	d.Copy(buff6[:])

	fmt.Printf("%x\n", buff6)
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

	if val, err := utf16.Bytes(buff6[:]); err == nil {
		fmt.Println(string(val))
	} else {
	}

	fmt.Println("eulaV")
}

type StringStructure struct {
}

func (s *StringStructure) Decode(d Decoder) {
	fmt.Println("string ptr", d.Uint32())
	fmt.Println("chiecksum", d.Uint16())
	count := d.Uint16()

	fmt.Println("count", count)

	buff6 := make([]byte, count*2+2)
	d.Copy(buff6[:])

	fmt.Printf("%x\n", buff6)
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

	if val, err := utf16.Bytes(buff6[:]); err == nil {
		fmt.Println(string(val))
	} else {
	}

	fmt.Println("gnirtS")
}

func Assert(b bool) {
	if b {
		return
	}
	panic("expr")
}

func main() {

	f, err := os.Open("/Users/remco/Projects/bosch/security.evtx")
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	index := 0

	fmt.Println(string(data[index : index+8]))
	index += 0x10

	d := NewDefaultDecoder(data, binary.LittleEndian)
	d.Seek(0x10)
	fmt.Println(d.Uint32())

	d.Seek(0x18)
	fmt.Println(d.Uint32())

	d.Seek(0x20)
	fmt.Println(d.Uint32())

	d.Seek(0x24)
	fmt.Println(d.Uint16())

	d.Seek(0x28)
	fmt.Println(d.Uint16())

	d.Seek(0x30)
	fmt.Println(d.Uint16())

	d.Seek(0x1000)

	buff := [8]byte{}
	d.Copy(buff[:])

	fmt.Println(string(buff[:]))

	d.Seek(0x1008)
	fmt.Println(d.Int64())
	d.Seek(0x1010)
	fmt.Println(d.Int64())
	d.Seek(0x1018)
	fmt.Println(d.Int64())
	d.Seek(0x1020)
	fmt.Println(d.Int64())

	d.Seek(0x1028)
	fmt.Println(d.Uint32())
	d.Seek(0x102c)
	fmt.Println(d.Uint32())
	d.Seek(0x1030)
	fmt.Println(d.Uint32())
	d.Seek(0x1034)
	fmt.Println(d.Uint32())

	d.Seek(0x1038)
	buff2 := [68]byte{}
	d.Copy(buff2[:])
	fmt.Println(string(buff2[:]))

	d.Seek(0x107c)
	fmt.Println(d.Uint32())

	d.Seek(0x1080)
	fmt.Println(d.Uint32())

	d.Seek(0x1180)
	fmt.Println(d.Uint32())
	/*
		0x000	char[8]	Magic, const 'ElfChnk', 0x00
		0x008	int64	NumLogRecFirst
		0x010	int64	NumLogRecLast
		0x018	int64	NumFileRecFirst
		0x020	int64	NumFileRecLast
		0x028	uint32	OfsTables, const 0x080
		0x02c	uint32	OfsRecLast
		0x030	uint32	OfsRecNext
		0x034	uint32	DataCRC
		0x038	char[68]	unknown
		0x07c	uint32	HeaderCRC
		0x080	uint32[64]	StringTable
		0x180	uint32[32]	TemplateTable
	*/

	d.Seek(0x1200)

	buff3 := [4]byte{}
	d.Copy(buff3[:])
	fmt.Println(string(buff3[:]))

	// length
	fmt.Println("Length", d.Uint32())

	// recordid
	fmt.Println("RecordID", d.Uint64())

	lowDateTime := int64(d.Uint32())
	highDateTime := int64(d.Uint32())
	nsec := int64(highDateTime)<<32 + int64(lowDateTime)
	nsec -= 116444736000000000
	nsec *= 100
	fmt.Println("Timestamp", time.Unix(0, nsec))

	if d.Uint8() == 0xf {
		fmt.Println("TEST")
		Assert(d.Uint8() == 0x1)
		fmt.Println("TEST1")
		Assert(d.Uint8() == 0x1)
		fmt.Println("TEST2")
		Assert(d.Uint8() == 0x0)
		fmt.Println("TEST3")
	}

	Assert(d.Uint8() == 0xc)
	Assert(d.Uint8() == 0x1)

	// https://static1.squarespace.com/static/510d93d8e4b060f86e6fdf2d/t/5328923ee4b0bea727f8aa9b/1395167806310/Windows+7+Audit+Format+v10.pdf
	fmt.Println("TemplateID", d.Uint32())
	fmt.Println("TemplatePointer", d.Uint32())

	// Template Definition
	fmt.Println("TemplatePointer", d.Uint32())
	buff5 := [16]byte{}
	d.Copy(buff5[:])
	fmt.Println("Guid", buff5)
	fmt.Println("Length", d.Uint32())
	fmt.Println("flag", d.Uint8())
	fmt.Println("unknown", d.Uint8())
	fmt.Println("unknown", d.Uint8())
	fmt.Println("unknown", d.Uint8())

	// element node
	en := &ElementNode{}
	en.Decode(d)

	return

	d.Seek(0x127a)
	fmt.Println(d.Uint16())

	d.Seek(0x1286)
	lowDateTime = int64(d.Uint32())
	highDateTime = int64(d.Uint32())
	nsec = int64(highDateTime)<<32 + int64(lowDateTime)
	nsec -= 116444736000000000
	nsec *= 100
	fmt.Println(time.Unix(0, nsec))

	d.Seek(0x1250)
	buff4 := [80]byte{}
	d.Copy(buff4[:])

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

	if val, err := utf16.Bytes(buff4[:]); err == nil {
		fmt.Println(string(val))
	} else {
	}

}
