package quic

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/h0n9/toybox/msg-lake/cluster"
	"github.com/quic-go/quic-go"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "quic",
	Short: "run quic server and client (lab)",
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterManagr := cluster.NewManager("0.0.0.0:8081")

		ctx, cancel := context.WithCancel(context.Background())
		wg := sync.WaitGroup{}

		// init sig channel
		sigCh := make(chan os.Signal, 1)
		defer close(sigCh)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// listen signals
		wg.Add(1)
		go func() {
			defer wg.Done()
			sig := <-sigCh
			fmt.Println("\r\ngot", sig.String())
			cancel()
			fmt.Printf("done\n")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := clusterManagr.Run(ctx)
			if err != nil {
				fmt.Println(err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			conn, err := quic.DialAddr("127.0.0.1:8081",
				&tls.Config{
					InsecureSkipVerify: true,
					NextProtos:         []string{"msg-lake"},
				},
				nil,
			)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer conn.CloseWithError(0, "bye")

			stream, err := conn.OpenStreamSync(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer stream.Close()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					_, err := stream.Write([]byte(fmt.Sprintf("%d\n", time.Now().UnixMilli())))
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}()

		wg.Wait()

		return nil
	},
}
