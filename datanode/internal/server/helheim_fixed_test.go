// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package server_test

import (
	pb "github.com/poy/loggrebutterfly/api/v1"
)

type mockWriteFetcher struct {
	WriterCalled chan bool
	WriterOutput struct {
		Writer chan func(data []byte) (err error)
		Err    chan error
	}
}

func newMockWriteFetcher() *mockWriteFetcher {
	m := &mockWriteFetcher{}
	m.WriterCalled = make(chan bool, 100)
	m.WriterOutput.Writer = make(chan func(data []byte) (err error), 100)
	m.WriterOutput.Err = make(chan error, 100)
	return m
}
func (m *mockWriteFetcher) Writer() (writer func(data []byte) (err error), err error) {
	m.WriterCalled <- true
	return <-m.WriterOutput.Writer, <-m.WriterOutput.Err
}

type mockReadFetcher struct {
	ReaderCalled chan bool
	ReaderInput  struct {
		Name       chan string
		StartIndex chan uint64
	}
	ReaderOutput struct {
		Reader chan func() (*pb.ReadData, error)
		Err    chan error
	}
}

func newMockReadFetcher() *mockReadFetcher {
	m := &mockReadFetcher{}
	m.ReaderCalled = make(chan bool, 100)
	m.ReaderInput.Name = make(chan string, 100)
	m.ReaderInput.StartIndex = make(chan uint64, 100)
	m.ReaderOutput.Reader = make(chan func() (*pb.ReadData, error), 100)
	m.ReaderOutput.Err = make(chan error, 100)
	return m
}
func (m *mockReadFetcher) Reader(name string, startIndex uint64) (reader func() (*pb.ReadData, error), err error) {
	m.ReaderCalled <- true
	m.ReaderInput.Name <- name
	m.ReaderInput.StartIndex <- startIndex
	return <-m.ReaderOutput.Reader, <-m.ReaderOutput.Err
}
