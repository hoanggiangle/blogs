package sgrpc

import (
	"context"
	"reflect"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestCopyMetadata(t *testing.T) {
	var ctx1, ctx2, dst context.Context
	var err error

	ctx1 = context.Background()
	ctx2 = context.Background()
	dst, err = CopyMetadata(ctx1, ctx2, "k1")
	if err != ErrorNotIncomingContext {
		t.Fatal("Wrong in context")
	}

	md := metadata.Pairs("k1", "v1", "k1", "v1'", "k2", "v2")
	ctx1 = metadata.NewIncomingContext(context.Background(), md)
	dst, err = CopyMetadata(ctx1, ctx2, "k1")
	if err != ErrorNotOutgoingContext {
		t.Fatal("Wrong out context")
	}

	ctx2 = metadata.NewOutgoingContext(context.Background(), metadata.Pairs())
	dst, err = CopyMetadata(ctx1, ctx2, "k1")

	if err != nil {
		t.Fatal(err)
	}

	md, _ = metadata.FromOutgoingContext(dst)
	if !reflect.DeepEqual(md, metadata.Pairs("k1", "v1", "k1", "v1'")) {
		t.Fatal("Wrong copy result")
	}
}
