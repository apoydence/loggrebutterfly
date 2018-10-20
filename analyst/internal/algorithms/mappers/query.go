package mappers

import (
	"strconv"

	loggregator "github.com/poy/loggrebutterfly/api/loggregator/v2"
	"github.com/golang/protobuf/proto"
)

type Filter interface {
	Filter(e *loggregator.Envelope) (keep bool)
}

type Query struct {
	filter Filter
}

func NewQuery(filter Filter) Query {
	return Query{
		filter: filter,
	}
}

func (r Query) Map(value []byte) (key string, output []byte, err error) {
	e, err := marshalAndFilter(value, r.filter)
	if err != nil || e == nil {
		return "", nil, err
	}

	return strconv.FormatInt(e.Timestamp, 10), value, nil
}

func marshalAndFilter(value []byte, filter Filter) (*loggregator.Envelope, error) {
	var e loggregator.Envelope
	if err := proto.Unmarshal(value, &e); err != nil {
		return nil, err
	}

	if filter.Filter(&e) {
		return &e, nil
	}

	return nil, nil
}
