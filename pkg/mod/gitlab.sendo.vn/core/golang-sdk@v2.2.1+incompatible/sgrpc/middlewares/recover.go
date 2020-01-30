package middlewares

import (
	"context"
	"fmt"
	"github.com/go-errors/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Recover() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				panicErr, ok := r.(error)
				if ok {
					fmt.Println(errors.Wrap(panicErr, 2).ErrorStack())
				} else {
					fmt.Println(r)
				}

				err = panicErr
				_ = status.Errorf(codes.Internal, "%s", panicErr)
			}
		}()
		return handler(ctx, req)
	}
}
