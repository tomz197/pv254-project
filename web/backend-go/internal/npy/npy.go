package npy

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

type Array struct {
	FortranOrder bool
	Shape        []int
	Dtype        string // e.g. "<f4", "<f8", "<i8", "<i4"
	Data         []byte // raw data payload
}

var magic = []byte{0x93, 'N', 'U', 'M', 'P', 'Y'}

func Read(r io.Reader) (*Array, error) {
	br := bufio.NewReader(r)
	head := make([]byte, 6)
	if _, err := io.ReadFull(br, head); err != nil { return nil, err }
	if !bytes.Equal(head, magic) { return nil, errors.New("not an npy file") }
	ver := make([]byte, 2)
	if _, err := io.ReadFull(br, ver); err != nil { return nil, err }
	var headerLen uint16
	if err := binary.Read(br, binary.LittleEndian, &headerLen); err != nil { return nil, err }
	hdr := make([]byte, headerLen)
	if _, err := io.ReadFull(br, hdr); err != nil { return nil, err }
	// header is a python dict like: {'descr': '<f4', 'fortran_order': False, 'shape': (N,M), }
	descr, fortran, shape, err := parseHeader(string(hdr))
	if err != nil { return nil, err }
	// compute data length requires element size and shape product
	esz, err := dtypeSize(descr)
	if err != nil { return nil, err }
	n := 1
	for _, d := range shape { n *= d }
	data := make([]byte, n*esz)
	if _, err := io.ReadFull(br, data); err != nil { return nil, err }
	return &Array{FortranOrder: fortran, Shape: shape, Dtype: descr, Data: data}, nil
}

func parseHeader(h string) (descr string, fortran bool, shape []int, err error) {
	// crude parsing; assumes well-formed header
	// find descr
	idx := bytes.Index([]byte(h), []byte("'descr':"))
	if idx < 0 { return "", false, nil, errors.New("no descr") }
	s := h[idx:]
	q1 := bytes.IndexByte([]byte(s), '\'')
	if q1 < 0 { q1 = 0 }
	q2 := bytes.Index([]byte(s[q1+1:]), []byte("'"))
	if q2 < 0 { return "", false, nil, errors.New("bad descr") }
	descr = s[q1+1 : q1+1+q2]
	// fortran_order
	fi := bytes.Index([]byte(h), []byte("fortran_order"))
	if fi >= 0 {
		fs := h[fi:]
		if bytes.Contains([]byte(fs), []byte("True")) { fortran = true }
	}
	// shape
	si := bytes.Index([]byte(h), []byte("shape"))
	if si < 0 { return "", false, nil, errors.New("no shape" ) }
	sb := []byte(h[si:])
	lp := bytes.IndexByte(sb, '(')
	rp := bytes.IndexByte(sb, ')')
	if lp < 0 || rp < 0 || rp <= lp { return "", false, nil, errors.New("bad shape") }
	inside := string(sb[lp+1 : rp])
	// split by comma
	parts := bytes.Split([]byte(inside), []byte{','})
	shape = make([]int, 0, len(parts))
	for _, p := range parts {
		ps := bytes.TrimSpace(p)
		if len(ps) == 0 { continue }
		var v int
		_, e := fmt.Sscanf(string(ps), "%d", &v)
		if e == nil { shape = append(shape, v) }
	}
	return descr, fortran, shape, nil
}

func dtypeSize(d string) (int, error) {
	switch d {
	case "<f8", "|f8": return 8, nil
	case "<f4", "|f4": return 4, nil
	case "<i8", "|i8": return 8, nil
	case "<i4", "|i4": return 4, nil
	default: return 0, fmt.Errorf("unsupported dtype %q", d)
	}
}

func (a *Array) Float32s() ([]float32, error) {
	switch a.Dtype {
	case "<f4", "|f4":
		out := make([]float32, len(a.Data)/4)
		for i := range out { out[i] = mathFloat32(binary.LittleEndian.Uint32(a.Data[i*4:])) }
		return out, nil
	case "<f8", "|f8":
		out := make([]float32, len(a.Data)/8)
		for i := range out { out[i] = float32(mathFloat64(binary.LittleEndian.Uint64(a.Data[i*8:]))) }
		return out, nil
	default:
		return nil, fmt.Errorf("dtype %s not convertible to float32", a.Dtype)
	}
}

func (a *Array) Ints() ([]int, error) {
	switch a.Dtype {
	case "<i8", "|i8":
		out := make([]int, len(a.Data)/8)
		for i := range out { out[i] = int(int64(binary.LittleEndian.Uint64(a.Data[i*8:]))) }
		return out, nil
	case "<i4", "|i4":
		out := make([]int, len(a.Data)/4)
		for i := range out { out[i] = int(int32(binary.LittleEndian.Uint32(a.Data[i*4:]))) }
		return out, nil
	default:
		return nil, fmt.Errorf("dtype %s not convertible to int", a.Dtype)
	}
}

func mathFloat32(u uint32) float32 { return *(*float32)(unsafe.Pointer(&u)) }
func mathFloat64(u uint64) float64 { return *(*float64)(unsafe.Pointer(&u)) }