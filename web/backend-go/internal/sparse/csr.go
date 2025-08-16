package sparse

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"

	"backendgo/internal/npy"
)

type CSR struct {
	Rows    int
	Cols    int
	Indptr  []int
	Indices []int
	Data    []float32
}

func LoadCSRFromNPZ(path string) (*CSR, error) {
	zr, err := zip.OpenReader(path)
	if err != nil { return nil, err }
	defer zr.Close()

	get := func(name string) ([]byte, error) {
		for _, f := range zr.File {
			if f.Name == name || f.Name == fmt.Sprintf("%s.npy", name) {
				rc, err := f.Open()
				if err != nil { return nil, err }
				defer rc.Close()
				var buf bytes.Buffer
				_, err = io.Copy(&buf, rc)
				return buf.Bytes(), err
			}
		}
		return nil, os.ErrNotExist
	}

	readFloat := func(b []byte) ([]float32, error) {
		a, err := npy.Read(bytes.NewReader(b))
		if err != nil { return nil, err }
		return a.Float32s()
	}
	readInt := func(b []byte) ([]int, error) {
		a, err := npy.Read(bytes.NewReader(b))
		if err != nil { return nil, err }
		return a.Ints()
	}

	dataB, err := get("data")
	if err != nil { return nil, err }
	indicesB, err := get("indices")
	if err != nil { return nil, err }
	indptrB, err := get("indptr")
	if err != nil { return nil, err }
	shapeB, err := get("shape")
	if err != nil { return nil, err }

	data, err := readFloat(dataB)
	if err != nil { return nil, err }
	indices, err := readInt(indicesB)
	if err != nil { return nil, err }
	indptr, err := readInt(indptrB)
	if err != nil { return nil, err }
	shape, err := readInt(shapeB)
	if err != nil { return nil, err }
	if len(shape) != 2 { return nil, fmt.Errorf("bad shape: %v", shape) }

	return &CSR{Rows: shape[0], Cols: shape[1], Indptr: indptr, Indices: indices, Data: data}, nil
}

// Row returns the non-zero indices and values for a given row
func (m *CSR) Row(i int) (idx []int, val []float32) {
	start := m.Indptr[i]
	end := m.Indptr[i+1]
	return m.Indices[start:end], m.Data[start:end]
}

// SumRows returns a dense float32 vector of length Cols with row sums of given indices.
func (m *CSR) SumRows(rows []int) []float32 {
	res := make([]float32, m.Cols)
	for _, r := range rows {
		idx, val := m.Row(r)
		for j, c := range idx {
			res[c] += val[j]
		}
	}
	return res
}