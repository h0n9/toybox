package lake

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/store"
	"github.com/stretchr/testify/assert"
)

func TestLake(t *testing.T) {
	// common
	addr := "localhost:8080"
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	done := make(chan bool)
	defer close(done)

	sampleMsgs := []*pb.Msg{}
	for i := 0; i < 1000; i++ {
		sampleMsgs = append(sampleMsgs, &pb.Msg{
			From: &pb.Address{Address: "addr-" + strconv.Itoa(rand.Int())},
			Data: &pb.Data{Data: []byte(strconv.Itoa(rand.Int()))},
		})
	}

	// init msgStore
	msgStore := store.NewMsgStoreMemory()

	// init server
	grpcServer := grpc.NewServer()
	lakeServer := NewLakeServer(msgStore)
	pb.RegisterLakeServer(grpcServer, lakeServer)
	listener, err := net.Listen("tcp", addr)
	assert.NoError(t, err)

	// run server
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = grpcServer.Serve(listener)
		assert.NoError(t, err)
	}()

	// init client
	connOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	connSender, err := grpc.Dial(addr, connOpts...)
	assert.NoError(t, err)
	connRecver, err := grpc.Dial(addr, connOpts...)
	assert.NoError(t, err)
	sender := pb.NewLakeClient(connSender)
	recver := pb.NewLakeClient(connRecver)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, sampleMsg := range sampleMsgs {
			res, err := sender.Send(
				ctx,
				&pb.SendReq{
					Id:  "test",
					Msg: sampleMsg,
				},
			)
			assert.NoError(t, err)
			assert.Equal(t, true, res.GetOk())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		stream, err := recver.Recv(ctx, &pb.RecvReq{Id: "test"})
		assert.NoError(t, err)
		defer stream.CloseSend()
		for i := range sampleMsgs {
			data, err := stream.Recv()
			assert.NoError(t, err)
			msg := data.GetMsg()
			assert.Equal(
				t,
				sampleMsgs[i].GetFrom().GetAddress(),
				msg.GetFrom().GetAddress(),
			)
			assert.Equal(
				t,
				sampleMsgs[i].GetData().GetData(),
				msg.GetData().GetData(),
			)
		}
		done <- true
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// block
		<-done
		cancel()

		// close server
		listener.Close()
		lakeServer.Close()
		grpcServer.GracefulStop()
	}()

	wg.Wait()
}
