package rcon

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

var (
	udpMarker = []byte{0xFF, 0xFF, 0xFF, 0xFF}
)

// Sender represents the remote console sender for Quake 3
type Sender interface {
	// Send sends command to the server and reads the response up to 64KB
	Send(ctx context.Context, command string) (output string, err error)
}

// NewSender creates a new instance of rcon command sender
func NewSender(dialAddr, password string) Sender {
	return &sender{
		dialAddr: dialAddr,
		password: password,
		buf:      make([]byte, 1024*64), // 64KB
	}
}

type sender struct {
	password string
	dialAddr string
	buf      []byte
}

func (s sender) Send(ctx context.Context, command string) (output string, err error) {
	logrus.Debugf("sending `%s` to `%s`...", command, s.dialAddr)
	logrus.Debugf("connecting to `%s`...", s.dialAddr)
	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", s.dialAddr)
	if err != nil {
		return "", err
	}
	logrus.Debugf("connection to `%s` created", s.dialAddr)
	defer func() {
		conn.Close()
		logrus.Debugf("connection to `%s` closed", s.dialAddr)
	}()

	msg := append(udpMarker, []byte(fmt.Sprintf("rcon %s %s\n", s.password, command))...)
	written, err := conn.Write(msg)
	if err != nil {
		return "", err
	}
	logrus.Debugf(
		"command `%s` has been sent to `%s`, %d bytes written",
		command,
		s.dialAddr,
		written,
	)

	logrus.Debugf("receiving response from `%s`", s.dialAddr)
	received, err := conn.Read(s.buf)
	logrus.Debugf("received %d bytes from `%s`", received, s.dialAddr)

	return string(s.buf[0:received]), err
}
