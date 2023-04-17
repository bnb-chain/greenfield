package client

import (
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield/sdk/keys"
)

// GreenfieldClientOption configures how we set up the greenfield client.
type GreenfieldClientOption interface {
	Apply(*GreenfieldClient)
}

// GreenfieldClientOptionFunc defines an applied function for setting the greenfield client.
type GreenfieldClientOptionFunc func(*GreenfieldClient)

// Apply set up the option field to the client instance.
func (f GreenfieldClientOptionFunc) Apply(client *GreenfieldClient) {
	f(client)
}

// WithKeyManager returns a GreenfieldClientOption which configures a client key manager option.
func WithKeyManager(km keys.KeyManager) GreenfieldClientOption {
	return GreenfieldClientOptionFunc(func(client *GreenfieldClient) {
		client.keyManager = km
	})
}

// WithGrpcConnectionAndDialOption returns a GreenfieldClientOption which configures a grpc client connection with grpc dail options.
func WithGrpcConnectionAndDialOption(grpcAddr string, opts ...grpc.DialOption) GreenfieldClientOption {
	return GreenfieldClientOptionFunc(func(client *GreenfieldClient) {
		client.grpcConn = grpcConn(grpcAddr, opts...)
	})
}

// grpcConn is used to establish a connection with a given address and dial options.
func grpcConn(addr string, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	return conn
}
