package azure

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/Azure/azure-storage-queue-go/azqueue"

	"github.com/tg123/xtcprelay"
)

type queueRelayer struct {
	queueClient *azqueue.ServiceURL
}

type queueRelayAddr struct {
	address string
}

func (a *queueRelayAddr) Network() string {
	return "xrelay-azure-stor-queue"
}

func (a *queueRelayAddr) String() string {
	return a.address
}

type queueConn struct {
	connaddr string

	sqinput  azqueue.MessagesURL
	sqoutput azqueue.MessagesURL

	outputbuf *bytes.Buffer

	close   bool
	onclose func() error
}

func (c *queueConn) Read(b []byte) (n int, err error) {

	for {
		if c.outputbuf.Len() > 0 {
			return c.outputbuf.Read(b)
		}

		messages, err := c.sqoutput.Dequeue(context.Background(), 10, 30*time.Second)

		if err != nil {
			return 0, err
		}

		if messages.NumMessages() == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for m := int32(0); m < messages.NumMessages(); m++ {
			msg := messages.Message(m)
			msgIDURL := c.sqoutput.NewMessageIDURL(msg.ID)
			_, err := msgIDURL.Delete(context.Background(), msg.PopReceipt)
			if err != nil {
				return 0, err
			}

			decoded, err := base64.StdEncoding.DecodeString(msg.Text)

			if err != nil {
				return 0, err
			}

			c.outputbuf.Write(decoded)
		}
	}
}

const MAX_SIZE = 8 << 10 // 8k (message body up to 64k)

func (c *queueConn) Write(b []byte) (int, error) {
	written := 0

	for i := 0; i < len(b); i += MAX_SIZE {
		if c.close {
			return written, io.EOF
		}

		s := len(b)
		if s > i+MAX_SIZE {
			s = i + MAX_SIZE
		}

		tmp := b[i:s]
		_, err := c.sqinput.Enqueue(context.Background(), base64.StdEncoding.EncodeToString(tmp), 0, -1*time.Second)
		if err != nil {
			return written, err
		}

		written += len(tmp)
	}

	return written, nil
}

func (c *queueConn) Close() error {
	c.close = true
	return c.onclose()
}

func (c *queueConn) LocalAddr() net.Addr {
	return &queueRelayAddr{c.connaddr}
}

func (c *queueConn) RemoteAddr() net.Addr {
	return &queueRelayAddr{c.connaddr}
}

func (c *queueConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *queueConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *queueConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type queueListener struct {
	address     string
	storequeue  azqueue.MessagesURL
	queueClient *azqueue.ServiceURL
	onclose     func() error
}

func (l *queueListener) Accept() (net.Conn, error) {
	for {

		messages, err := l.storequeue.Dequeue(context.Background(), 1, 10*time.Second)

		if err != nil {
			return nil, err
		}

		if messages.NumMessages() == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		msg := messages.Message(0)
		connid := msg.Text
		msgIDURL := l.storequeue.NewMessageIDURL(msg.ID)
		msgIDURL.Delete(context.Background(), msg.PopReceipt)

		in := l.queueClient.NewQueueURL(fmt.Sprintf("%v-in", connid))
		out := l.queueClient.NewQueueURL(fmt.Sprintf("%v-out", connid))

		return &queueConn{
			connaddr:  l.address,
			sqinput:   in.NewMessagesURL(),
			sqoutput:  out.NewMessagesURL(),
			outputbuf: bytes.NewBuffer(nil),
			onclose: func() error {
				in.Delete(context.Background())
				out.Delete(context.Background())
				return nil
			},
		}, nil

	}
}

func (l *queueListener) Close() error {
	return l.onclose()
}

func (l *queueListener) Addr() net.Addr {
	return &queueRelayAddr{l.address}
}

func (r *queueRelayer) Listen(address string) (net.Listener, error) {
	queue := r.queueClient.NewQueueURL(fmt.Sprintf("%v-socket", address))
	_, err := queue.Create(context.Background(), azqueue.Metadata{})
	if err != nil {
		return nil, err
	}
	return &queueListener{
		storequeue:  queue.NewMessagesURL(),
		queueClient: r.queueClient,
		onclose: func() error {
			_, err := queue.Delete(context.Background())
			return err
		},
	}, nil
}

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (r *queueRelayer) Dial(address string) (net.Conn, error) {

	connid := randStringBytes(10)

	in := r.queueClient.NewQueueURL(fmt.Sprintf("%v-in", connid))
	_, err := in.Create(context.Background(), azqueue.Metadata{})
	if err != nil {
		return nil, err
	}

	out := r.queueClient.NewQueueURL(fmt.Sprintf("%v-out", connid))
	_, err = out.Create(context.Background(), azqueue.Metadata{})
	if err != nil {
		return nil, err
	}

	queue := r.queueClient.NewQueueURL(fmt.Sprintf("%v-socket", address)).NewMessagesURL()

	_, err = queue.Enqueue(context.Background(), connid, 0, -1*time.Second)
	if err != nil {
		return nil, err
	}

	return &queueConn{
		connaddr:  address,
		sqinput:   out.NewMessagesURL(),
		sqoutput:  in.NewMessagesURL(),
		outputbuf: bytes.NewBuffer(nil),
		onclose: func() error {
			in.Delete(context.Background())
			out.Delete(context.Background())
			return nil
		},
	}, nil
}

func (r *queueRelayer) Close() error {
	return nil
}

func newQueueRelayer(queue *azqueue.ServiceURL) (*queueRelayer, error) {
	return &queueRelayer{
		queueClient: queue,
	}, nil
}

func NewQueueRelayerClientSide(queue *azqueue.ServiceURL) (xtcprelay.RelayerClientSide, error) {
	return newQueueRelayer(queue)
}

func NewQueueRelayerServerSide(queue *azqueue.ServiceURL) (xtcprelay.RelayerServerSide, error) {
	return newQueueRelayer(queue)
}
