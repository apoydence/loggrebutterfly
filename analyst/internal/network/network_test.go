package network_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/apoydence/loggrebutterfly/analyst/internal/network"
	"github.com/apoydence/loggrebutterfly/api/intra"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	os.Exit(m.Run())
}

type TN struct {
	*testing.T
	n                 *network.Network
	mockAnalystServer *mockAnalystServer
	analystAddr       string
}

func TestNetwork(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TN {
		addr, mockAnalystServer := startAnalystServer()

		return TN{
			T:                 t,
			n:                 network.New(),
			mockAnalystServer: mockAnalystServer,
			analystAddr:       addr,
		}
	})

	o.Group("when the server does not return an error", func() {
		o.BeforeEach(func(t TN) TN {
			close(t.mockAnalystServer.ExecuteOutput.Ret1)
			t.mockAnalystServer.ExecuteOutput.Ret0 <- &intra.ExecuteResponse{
				Result: map[string][]byte{
					"a": []byte("some-a"),
					"b": []byte("some-b"),
				},
			}
			return t
		})

		o.Spec("it returns the results", func(t TN) {
			results, err := t.n.Execute("some-file", "some-alg", t.analystAddr, context.Background(), nil)
			Expect(t, err == nil).To(BeTrue())
			Expect(t, results).To(HaveLen(2))
			Expect(t, results["a"]).To(Equal([]byte("some-a")))
			Expect(t, results["b"]).To(Equal([]byte("some-b")))
		})

		o.Spec("it executes with the expected file, alg and meta", func(t TN) {
			t.n.Execute("some-file", "some-alg", t.analystAddr, context.Background(), []byte("meta"))

			Expect(t, t.mockAnalystServer.ExecuteInput.Arg1).To(
				Chain(Receive(), Equal(&intra.ExecuteInfo{
					File: "some-file",
					Alg:  "some-alg",
					Meta: []byte("meta"),
				})),
			)
		})
	})

	o.Group("when the server returns an error", func() {
		o.BeforeEach(func(t TN) TN {
			close(t.mockAnalystServer.ExecuteOutput.Ret0)
			t.mockAnalystServer.ExecuteOutput.Ret1 <- fmt.Errorf("some-error")
			return t
		})

		o.Spec("it returns an error", func(t TN) {
			_, err := t.n.Execute("some-file", "some-alg", t.analystAddr, context.Background(), nil)
			Expect(t, err == nil).To(BeFalse())
		})
	})
}

func startAnalystServer() (string, *mockAnalystServer) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	mockAnalystServer := newMockAnalystServer()
	intra.RegisterAnalystServer(s, mockAnalystServer)
	go s.Serve(lis)

	return lis.Addr().String(), mockAnalystServer
}
