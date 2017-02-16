package reducers

import (
	"encoding/binary"
	"fmt"
	"math"
)

type SumF struct {
}

func NewSumF() SumF {
	return SumF{}
}

func (s SumF) Reduce(value [][]byte) ([][]byte, error) {
	var total float64
	for _, x := range value {
		if len(x) != 8 {
			return nil, fmt.Errorf("not a float64: %v", x)
		}
		f := bytesToFloat(x)
		total += f
	}

	return [][]byte{floatToBytes(total)}, nil
}

func floatToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func bytesToFloat(b []byte) float64 {
	bits := binary.LittleEndian.Uint64(b)
	float := math.Float64frombits(bits)
	return float
}
