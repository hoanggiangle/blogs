package sgrpc

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func GetClientIp(ctx context.Context) net.IP {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}
	switch p.Addr.(type) {
	case *net.TCPAddr:
		return p.Addr.(*net.TCPAddr).IP
	case *net.UDPAddr:
		return p.Addr.(*net.UDPAddr).IP
	}
	return nil
}

func findCommonPrefix(keys []string) int {
	var commonPrefix = -1

	if len(keys) == 0 {
		return -1
	} else if len(keys) == 1 {
		return len(keys[0])
	}

DONE_CP:
	for i := 0; i < len(keys[0]); i++ {
		for _, k := range keys[1:] {
			if keys[0][i] != k[i] {
				break DONE_CP
			}
		}
		commonPrefix = i
	}

	return commonPrefix
}

var (
	ErrorNotIncomingContext = errors.New("Not is incoming context")
	ErrorNotOutgoingContext = errors.New("Not is outgoing context")
)

// copy metadata from incoming context into outgoing context
func CopyMetadata(src context.Context, into context.Context, metakeys ...string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(src)
	if !ok {
		return into, ErrorNotIncomingContext
	}
	if _, ok = metadata.FromOutgoingContext(into); !ok {
		return into, ErrorNotOutgoingContext
	}

	args := make([]string, 0)
	for _, k := range metakeys {
		if val := md.Get(k); len(val) > 0 {
			for _, v := range val {
				args = append(args, k, v)
			}
		}
	}
	if len(args) == 0 {
		return into, nil
	}
	return metadata.AppendToOutgoingContext(into, args...), nil
}
