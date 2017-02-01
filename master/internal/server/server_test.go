//go:generate hel

package server_test

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/apoydence/loggrebutterfly/master/internal/server"
	pb "github.com/apoydence/loggrebutterfly/pb/v1"
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

type TS struct {
	*testing.T
	masterClient pb.MasterClient
	mockLister   *mockLister
}

func TestServer(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TS {
		mockLister := newMockLister()
		addr, err := server.Start("127.0.0.1:0", mockLister)
		Expect(t, err == nil).To(BeTrue())

		return TS{
			T:            t,
			masterClient: fetchClient(addr),
			mockLister:   mockLister,
		}
	})

	o.Spec("it returns the results from the lister", func(t TS) {
		close(t.mockLister.RoutesOutput.Err)
		t.mockLister.RoutesOutput.Routes <- map[string]string{
			"some-route-a": "some-leader",
			"some-route-b": "some-leader",
		}

		ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		resp, err := t.masterClient.Routes(ctx, new(pb.RoutesInfo))
		Expect(t, err == nil).To(BeTrue())
		Expect(t, resp.Routes).To(HaveLen(2))
		Expect(t, resp.Routes[0].Name).To(Or(
			Equal("some-route-a"),
			Equal("some-route-b"),
		))
		Expect(t, resp.Routes[1].Name).To(Or(
			Equal("some-route-a"),
			Equal("some-route-b"),
		))
		Expect(t, resp.Routes[0].Name).To(Not(Equal(resp.Routes[1].Name)))
		Expect(t, resp.Routes[0].Leader).To(Equal("some-leader"))
	})
}

func fetchClient(addr string) pb.MasterClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewMasterClient(conn)
}
