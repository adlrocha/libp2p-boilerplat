package server

import (
	"context"
	"io"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-msgio"

	"github.com/adlrocha/libp2p-msg/req-resp-example/net"
	pb "github.com/adlrocha/libp2p-msg/req-resp-example/pb"
)

// Idle time before the stream is closed
var streamIdleTimeout = 1 * time.Minute
var _ Server = &server{}

type Server interface {
	setProtocolHandler(network.StreamHandler)
}

// SmartRecordServer handles smart-record requests
type server struct {
	ctx       context.Context
	host      host.Host
	self      peer.ID
	protocols []protocol.ID
}

func (e *server) setProtocolHandler(h network.StreamHandler) {
	// For every announce protocol set this new handler.
	for _, p := range e.protocols {
		e.host.SetStreamHandler(p, h)
	}
}

func NewServer(ctx context.Context, h host.Host, options ...ServerOption) (*server, error) {
	var cfg serverConfig
	if err := cfg.apply(append([]ServerOption{serverDefaults}, options...)...); err != nil {
		return nil, err
	}
	protocols := []protocol.ID{pid}

	e := &server{
		ctx:       ctx,
		host:      h,
		self:      h.ID(),
		protocols: protocols,
	}

	e.setProtocolHandler(e.handleNewStream)

	return e, nil
}

type handler func(context.Context, peer.ID, *pb.Message) (*pb.Message, error)

func (e *server) handlerForMsgType(t pb.Message_MessageType) handler {
	switch t {
	case pb.Message_GET:
		return e.handleGet
		// TODO: Add here more message handlers. You will also need to add
		// new types to the protobuf
	}

	return nil
}

func (e *server) handleGet(ctx context.Context, p peer.ID, msg *pb.Message) (*pb.Message, error) {

	// setup response with same type as request.
	// TODO: right now the handler returns the same message.
	// Implement here the logic you want for the handler.
	resp := &pb.Message{
		Type:  msg.GetType(),
		Value: msg.GetValue(),
	}

	return resp, nil
}

// handleNewStream implements the network.StreamHandler
func (e *server) handleNewStream(s network.Stream) {
	if e.handleNewMessages(s) {
		// If we exited without error, close gracefully.
		_ = s.Close()
	} else {
		// otherwise, send an error.
		_ = s.Reset()
	}
}

// Returns true on orderly completion of writes (so we can Close the stream conveniently).
func (e *server) handleNewMessages(s network.Stream) bool {
	ctx := e.ctx
	r := msgio.NewVarintReaderSize(s, network.MessageSizeMax)

	mPeer := s.Conn().RemotePeer()

	timer := time.AfterFunc(streamIdleTimeout, func() { _ = s.Reset() })
	defer timer.Stop()

	for {
		var req pb.Message
		msgbytes, err := r.ReadMsg()
		if err != nil {
			r.ReleaseMsg(msgbytes)
			if err == io.EOF {
				return true
			}
			return false
		}
		err = req.Unmarshal(msgbytes)
		r.ReleaseMsg(msgbytes)
		if err != nil {
			return false
		}

		timer.Reset(streamIdleTimeout)

		handler := e.handlerForMsgType(req.GetType())
		if handler == nil {
			return false
		}

		resp, err := handler(ctx, mPeer, &req)
		if err != nil {
			return false
		}

		if resp == nil {
			continue
		}

		// send out response msg
		err = net.WriteMsg(s, resp)
		if err != nil {
			return false
		}

	}
}
