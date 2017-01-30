package hasher

import (
	"hash/fnv"

	v2 "github.com/apoydence/loggrebutterfly/pb/loggregator/v2"
	"github.com/golang/protobuf/proto"
)

type Hasher struct {
}

func New() *Hasher {
	return &Hasher{}
}

func (h *Hasher) HashString(s string) (hash uint64) {
	f := fnv.New64a()
	f.Write([]byte(s))

	return f.Sum64()
}

// Hash takes an envelope and converts it's SourceUUID into a hash
func (h *Hasher) Hash(data []byte) (hash uint64, err error) {
	var e v2.Envelope
	if err := proto.Unmarshal(data, &e); err != nil {
		return 0, err
	}

	return h.HashString(e.SourceUuid), nil
}
