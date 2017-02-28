// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package mappers_test

import (
	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
)

type mockFilter struct {
	FilterCalled chan bool
	FilterInput  struct {
		E chan *loggregator.Envelope
	}
	FilterOutput struct {
		Keep chan bool
	}
}

func newMockFilter() *mockFilter {
	m := &mockFilter{}
	m.FilterCalled = make(chan bool, 100)
	m.FilterInput.E = make(chan *loggregator.Envelope, 100)
	m.FilterOutput.Keep = make(chan bool, 100)
	return m
}
func (m *mockFilter) Filter(e *loggregator.Envelope) (keep bool) {
	m.FilterCalled <- true
	m.FilterInput.E <- e
	return <-m.FilterOutput.Keep
}
