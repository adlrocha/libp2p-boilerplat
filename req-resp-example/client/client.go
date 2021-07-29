package client

import (
	"context"

	pb "github.com/adlrocha/libp2p-msg/req-resp-example/pb"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

var log = logging.Logger("client")
var _ Client = &client{}

type Client interface {
	SendGet(ctx context.Context, value []byte, p peer.ID) (interface{}, error)
}

// smartRecordClient is responsible for sending smart-record
// requests to other peers.
type client struct {
	ctx       context.Context
	host      host.Host
	self      peer.ID
	protocols []protocol.ID

	senderManager *messageSenderImpl
}

func NewClient(ctx context.Context, h host.Host, options ...ClientOption) (*client, error) {
	var cfg clientConfig
	if err := cfg.apply(append([]ClientOption{clientDefaults}, options...)...); err != nil {
		return nil, err
	}
	protocols := []protocol.ID{pid}

	// Start a smartRecordClient
	e := &client{
		ctx:       ctx,
		host:      h,
		self:      h.ID(),
		protocols: protocols,

		senderManager: &messageSenderImpl{
			host:      h,
			strmap:    make(map[peer.ID]*peerMessageSender),
			protocols: protocols,
		},
	}

	return e, nil
}

func (e *client) SendGet(ctx context.Context, value []byte, p peer.ID) (interface{}, error) {
	// Send a new request and wait for response
	// TODO: Add the request you want to send to the server here
	req := &pb.Message{
		Type:  pb.Message_GET,
		Value: value,
	}
	resp, err := e.senderManager.SendRequest(ctx, p, req)
	if err != nil {
		return nil, err
	}

	// Return the response received without doing anything else
	return resp.GetValue(), nil

}
