// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package intra_test

import "golang.org/x/net/context"

type mockExecutor struct {
	ExecuteCalled chan bool
	ExecuteInput  struct {
		FileName, AlgName chan string
		Ctx               chan context.Context
		Meta              chan []byte
	}
	ExecuteOutput struct {
		Result chan map[string][]byte
		Err    chan error
	}
}

func newMockExecutor() *mockExecutor {
	m := &mockExecutor{}
	m.ExecuteCalled = make(chan bool, 100)
	m.ExecuteInput.FileName = make(chan string, 100)
	m.ExecuteInput.AlgName = make(chan string, 100)
	m.ExecuteInput.Ctx = make(chan context.Context, 100)
	m.ExecuteInput.Meta = make(chan []byte, 100)
	m.ExecuteOutput.Result = make(chan map[string][]byte, 100)
	m.ExecuteOutput.Err = make(chan error, 100)
	return m
}
func (m *mockExecutor) Execute(fileName, algName string, ctx context.Context, meta []byte) (result map[string][]byte, err error) {
	m.ExecuteCalled <- true
	m.ExecuteInput.FileName <- fileName
	m.ExecuteInput.AlgName <- algName
	m.ExecuteInput.Ctx <- ctx
	m.ExecuteInput.Meta <- meta
	return <-m.ExecuteOutput.Result, <-m.ExecuteOutput.Err
}