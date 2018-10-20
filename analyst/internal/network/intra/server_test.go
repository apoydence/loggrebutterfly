package intra_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/poy/loggrebutterfly/analyst/internal/network/intra"
	api "github.com/poy/loggrebutterfly/api/intra"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
)

//go:generate hel

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	os.Exit(m.Run())
}

type TI struct {
	*testing.T
	s        *intra.Server
	mockExec *mockExecutor
}

func TestIntraServer(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TI {
		mockExec := newMockExecutor()
		return TI{
			T:        t,
			s:        intra.New(mockExec),
			mockExec: mockExec,
		}
	})

	o.Group("when the executor does not return an error", func() {
		o.BeforeEach(func(t TI) TI {
			close(t.mockExec.ExecuteOutput.Err)
			t.mockExec.ExecuteOutput.Result <- map[string][]byte{
				"a": []byte("value-a"),
				"b": []byte("value-b"),
			}
			return t
		})

		o.Spec("it returns the results", func(t TI) {
			resp, err := t.s.Execute(context.Background(), &api.ExecuteInfo{File: "f", Alg: "a"})
			Expect(t, err == nil).To(BeTrue())
			Expect(t, resp.Result).To(HaveLen(2))
			Expect(t, resp.Result["a"]).To(Equal([]byte("value-a")))
			Expect(t, resp.Result["b"]).To(Equal([]byte("value-b")))
		})

		o.Spec("it returns an error for an empty file and alg", func(t TI) {
			_, err := t.s.Execute(context.Background(), &api.ExecuteInfo{Alg: "a"})
			Expect(t, err == nil).To(BeFalse())
			_, err = t.s.Execute(context.Background(), &api.ExecuteInfo{File: "f"})
			Expect(t, err == nil).To(BeFalse())
		})

		o.Spec("it gives the executor the correct args", func(t TI) {
			t.s.Execute(context.Background(), &api.ExecuteInfo{File: "f", Alg: "a", Meta: []byte("meta")})

			Expect(t, t.mockExec.ExecuteInput.FileName).To(
				Chain(Receive(), Equal("f")),
			)
			Expect(t, t.mockExec.ExecuteInput.AlgName).To(
				Chain(Receive(), Equal("a")),
			)
			Expect(t, t.mockExec.ExecuteInput.Meta).To(
				Chain(Receive(), Equal([]byte("meta"))),
			)
		})
	})

	o.Group("when the executor returns an error", func() {
		o.BeforeEach(func(t TI) TI {
			close(t.mockExec.ExecuteOutput.Result)
			t.mockExec.ExecuteOutput.Err <- fmt.Errorf("some-error")
			return t
		})

		o.Spec("it returns an error", func(t TI) {
			_, err := t.s.Execute(context.Background(), &api.ExecuteInfo{File: "f", Alg: "a"})
			Expect(t, err == nil).To(BeFalse())
		})
	})

}
