package reducers

type First struct {
}

func NewFirst() First {
	return First{}
}

func (f First) Reduce(value [][]byte) ([][]byte, error) {
	if len(value) == 0 {
		return nil, nil
	}

	return [][]byte{value[0]}, nil
}
