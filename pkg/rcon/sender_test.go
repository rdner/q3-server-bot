package rcon

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// logrus.SetOutput(ioutil.Discard)

	addr := runEchoServer(ctx, t)

	cases := []struct {
		name      string
		dialAddr  string
		password  string
		command   string
		expError  string
		expOutput string
	}{
		{
			name:      "successfully sends the command if the configuration is valid",
			dialAddr:  addr,
			password:  "password",
			command:   "command",
			expOutput: "rcon password command\n",
		},
		{
			name:     "returns error if server not found",
			dialAddr: ":8888",
			password: "password",
			command:  "command",
			expError: "connection refused",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sender := NewSender(tc.dialAddr, tc.password)
			output, err := sender.Send(ctx, tc.command)
			if tc.expError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expOutput, output)
		})
	}
}

func runEchoServer(ctx context.Context, t *testing.T) (addr string) {
	pc, err := net.ListenPacket("udp", "0.0.0.0:0")
	require.NoError(t, err)
	err = pc.SetDeadline(time.Now().Add(30 * time.Second))
	require.NoError(t, err)

	go func() {
		defer pc.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				message := make([]byte, 1024)
				rlen, remote, err := pc.ReadFrom(message[:])
				require.NoError(t, err)
				_, err = pc.WriteTo(message[len(udpMarker):rlen], remote)
				require.NoError(t, err)
			}
		}
	}()

	return pc.LocalAddr().String()
}
