package embeddings

import (
	"fmt"
	"os"

	"backendgo/internal/npy"
)

type Matrix struct {
	Rows int
	Cols int
	Data []float32 // row-major flattened
}

func LoadNPY(path string) (*Matrix, error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()
	arr, err := npy.Read(f)
	if err != nil { return nil, err }
	if len(arr.Shape) != 2 {
		return nil, fmt.Errorf("expected 2D array, got %v", arr.Shape)
	}
	vals, err := arr.Float32s()
	if err != nil { return nil, err }
	return &Matrix{Rows: arr.Shape[0], Cols: arr.Shape[1], Data: vals}, nil
}

func (m *Matrix) Row(i int) []float32 {
	start := i * m.Cols
	return m.Data[start : start+m.Cols]
}