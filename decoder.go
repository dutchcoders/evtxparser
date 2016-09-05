package evtxparser

import (
	"encoding/binary"
	"fmt"
	"math"
)

type ErrDecoderTooShort struct {
	Got  int
	Want int
}

func (e ErrDecoderTooShort) Error() string {
	return fmt.Sprintf("DCERPC Layer decoding: length %v too short, %v required", e.Got, e.Want)
}

type Decoder interface {
	HasBytes(size int) bool

	PeekUint8() uint8
	PeekUint16() uint16

	Byte() byte

	Int8() int8
	Int16() int16
	Int32() int32
	Int64() int64

	Uint8() uint8
	Uint16() uint16
	Uint32() uint32
	Uint64() uint64

	IEEE754_Float32() float32
	IEEE754_Float64() float64

	CString() string
	Data() []byte

	Copy([]byte)

	LastError() error
	SetLastError(error)

	Skip(int)
	Align(int)
	Seek(int) int

	StartOffset() int
	Offset() int

	NewDecoder() Decoder

	Dump()

	ByteOrder() binary.ByteOrder
	SetByteOrder(byteOrder binary.ByteOrder) binary.ByteOrder
}

func NewDefaultDecoder(data []byte, byteOrder binary.ByteOrder) Decoder {
	return &DefaultDecoder{
		offset:      0,
		startOffset: 0,
		byteOrder:   byteOrder,
		data:        data,
	}
}

type DefaultDecoder struct {
	offset      int
	startOffset int
	lastError   error
	byteOrder   binary.ByteOrder
	data        []byte
}

func (d *DefaultDecoder) NewDecoder() Decoder {
	return &DefaultDecoder{
		offset:      0,
		startOffset: 0,
		byteOrder:   d.byteOrder,
		data:        d.data[d.offset:],
	}
}

func (d *DefaultDecoder) SetByteOrder(byteOrder binary.ByteOrder) binary.ByteOrder {
	prevByteOrder := d.byteOrder
	d.byteOrder = byteOrder
	return prevByteOrder
}

func (d *DefaultDecoder) ByteOrder() binary.ByteOrder {
	return d.byteOrder

}

func (d *DefaultDecoder) Offset() int {
	return d.offset

}

func (d *DefaultDecoder) StartOffset() int {
	return d.startOffset

}

func (d *DefaultDecoder) HasBytes(size int) bool {
	if len(d.data) >= d.offset+size {
		return true
	}

	d.lastError = ErrDecoderTooShort{
		Got:  size,
		Want: len(d.data) - d.offset,
	}

	return false
}

// not special this guid
func (d *DefaultDecoder) Copy(buff []byte) {
	if d.lastError != nil {
		return
	}

	if !d.HasBytes(len(buff)) {
		return
	}

	defer func() {
		d.offset += len(buff)
	}()

	copy(buff[:], d.data[d.offset:d.offset+len(buff)])
	return
}

func (d *DefaultDecoder) Byte() byte {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(1) {
		return 0
	}

	defer func() {
		d.offset += 1
	}()

	return d.data[d.offset]
}

func (d *DefaultDecoder) PeekUint8() uint8 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(1) {
		return 0
	}

	return uint8(d.data[d.offset])
}

func (d *DefaultDecoder) Int8() int8 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(1) {
		return 0
	}

	defer func() {
		d.offset += 1
	}()

	return int8(uint8(d.data[d.offset]))
}

func (d *DefaultDecoder) Uint8() uint8 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(1) {
		return 0
	}

	defer func() {
		d.offset += 1
	}()

	return uint8(d.data[d.offset])
}

func (d *DefaultDecoder) Int16() int16 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(2) {
		return 0
	}

	defer func() {
		d.offset += 2
	}()

	return int16(d.byteOrder.Uint16(d.data[d.offset : d.offset+2]))
}

func (d *DefaultDecoder) Uint16() uint16 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(2) {
		return 0
	}

	defer func() {
		d.offset += 2
	}()

	return d.byteOrder.Uint16(d.data[d.offset : d.offset+2])
}

func (d *DefaultDecoder) PeekUint16() uint16 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(2) {
		return 0
	}

	return d.byteOrder.Uint16(d.data[d.offset : d.offset+2])
}

func (d *DefaultDecoder) Seek(n int) int {
	prev := d.offset
	d.offset = n
	return prev
}

/*
func (d *DefaultDecoder) Offset() int {
	return d.offset
}
*/

func (d *DefaultDecoder) Align(n int) {
	if d.offset%n == 0 {
		return
	}

	d.offset += n - (d.offset % n)
}

func (d *DefaultDecoder) Skip(n int) {
	d.offset += n
}

// TODO: make efficient
func (d *DefaultDecoder) CString() string {
	str := ""
	for !(d.data[d.offset] == 0x00 && d.data[d.offset+1] == 0x00) {
		str += string(d.data[d.offset : d.offset+1])
		d.offset += 2
	}

	d.offset += 2
	return str
}

func (d *DefaultDecoder) Int32() int32 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(4) {
		return 0
	}

	defer func() {
		d.offset += 4
	}()

	return int32(d.byteOrder.Uint32(d.data[d.offset : d.offset+4]))
}

func (d *DefaultDecoder) Uint32() uint32 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(4) {
		return 0
	}

	defer func() {
		d.offset += 4
	}()

	return d.byteOrder.Uint32(d.data[d.offset : d.offset+4])
}

func (d *DefaultDecoder) IEEE754_Float32() float32 {
	if d.lastError != nil {
		return 0
	}

	v := d.Uint32()
	return math.Float32frombits(v)
}

func (d *DefaultDecoder) IEEE754_Float64() float64 {
	if d.lastError != nil {
		return 0
	}

	v := d.Uint64()
	return math.Float64frombits(v)
}

func (d *DefaultDecoder) Int64() int64 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(8) {
		return 0
	}

	defer func() {
		d.offset += 8
	}()

	return int64(d.byteOrder.Uint64(d.data[d.offset : d.offset+8]))
}

func (d *DefaultDecoder) Uint64() uint64 {
	if d.lastError != nil {
		return 0
	}

	if !d.HasBytes(8) {
		return 0
	}

	defer func() {
		d.offset += 8
	}()

	return d.byteOrder.Uint64(d.data[d.offset : d.offset+8])
}

func (d *DefaultDecoder) Data() []byte {
	return d.data[:]
}

/*
func (d *DefaultDecoder) SetData(err error) {
	d.lastError = err
}
*/

func (d *DefaultDecoder) LastError() error {
	return d.lastError
}

func (d *DefaultDecoder) SetLastError(err error) {
	d.lastError = err
}

func (d *DefaultDecoder) Dump() {
	// debug.PrintStack()

	fmt.Printf("Offset: %d (%x)\n% #x \n", d.offset, d.offset, d.data[d.offset:d.offset+100])
}
