package hasher

import (
	"hash/fnv"

	v2 "github.com/poy/loggrebutterfly/api/loggregator/v2"
	"github.com/golang/protobuf/proto"
)

type Hasher struct {
}

func New() *Hasher {
	return &Hasher{}
}

func (h *Hasher) Hash(data []byte) (hash uint64, err error) {
	var e v2.Envelope
	if err := proto.Unmarshal(data, &e); err != nil {
		return 0, err
	}

	f := fnv.New64a()
	f.Write([]byte(e.SourceId))

	return f.Sum64(), nil
}
