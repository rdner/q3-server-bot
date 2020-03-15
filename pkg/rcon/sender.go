package rcon

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

var (
	udpMarker = []byte{0xFF, 0xFF, 0xFF, 0xFF}
)

type Sender interface {
	Send(ctx context.Context, command string) (output string, err error)
}

func NewSender(dialAddr, password string) Sender {
	return &sender{
		dialAddr: dialAddr,
		password: password,
	}
}

type sender struct {
	password string
	dialAddr string
}

func (s sender) Send(ctx context.Context, command string) (output string, err error) {
	logrus.Debugf("sending `%s` to `%s`...", command, s.dialAddr)
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
	bts, err := scanResponse(ctx, conn)
	logrus.Debugf("received %d bytes from `%s`", len(bts), s.dialAddr)

	return string(bts), err
}

func scanResponse(ctx context.Context, conn io.Reader) (output string, err error) {
	scanner := bufio.NewScanner(conn)
	scanner.Split(splitByMarkerFunc)

	if !scanner.Scan() {
		return "", errors.New("scanned no response")
	}
	return scanner.Text(), nil
}

func splitByMarkerFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	mi := bytes.Index(data, udpMarker)
	if mi == -1 {
		return 0, nil, nil
	}
	return mi + len(udpMarker), data[0:mi], nil
}
