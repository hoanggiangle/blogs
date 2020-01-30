package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"regexp"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"gitlab.sendo.vn/core/golang-sdk/slog"
)

type LogCondition interface {
	ShouldLogUnary(method string, req interface{}) bool
	ShouldLogStream(method string) bool
}

type allwayCondition struct{}

func (c *allwayCondition) ShouldLogUnary(method string, req interface{}) bool {
	return true
}
func (c *allwayCondition) ShouldLogStream(method string) bool {
	return true
}

type simpleCondition struct {
	methods []string
}

func (c *simpleCondition) ShouldLogUnary(method string, req interface{}) bool {
	return c.ShouldLogStream(method)
}
func (c *simpleCondition) ShouldLogStream(method string) bool {
	for _, m := range c.methods {
		if method == m {
			return true
		}
	}
	return false
}

func NewSimpleLogCondition(methods []string) LogCondition {
	return &simpleCondition{methods}
}

type regexCondition struct {
	re *regexp.Regexp
}

func (c *regexCondition) ShouldLogUnary(method string, req interface{}) bool {
	return c.re.MatchString(method)
}
func (c *regexCondition) ShouldLogStream(method string) bool {
	return c.re.MatchString(method)
}

func NewRegexLogCondition(re *regexp.Regexp) LogCondition {
	return &regexCondition{re}
}

func getRequestId(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if reqId := md.Get("x-request-id"); len(reqId) > 0 {
			return reqId[0]
		}
	}

	b := make([]byte, 12)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func grpcErrorToJson(err error) string {
	r := map[string]string{
		"status":  grpc.Code(err).String(),
		"message": grpc.ErrorDesc(err),
	}

	b, _ := json.Marshal(r)
	return string(b)
}

func NewLoggingUnaryServerInterceptor(cond LogCondition, logger slog.Logger) grpc.UnaryServerInterceptor {
	var marshaler = jsonpb.Marshaler{OrigName: true}

	if cond == nil {
		cond = &allwayCondition{}
	}

	f := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		resp, err = handler(ctx, req)

		if !cond.ShouldLogUnary(info.FullMethod, req) {
			return
		}

		fields := slog.Fields{
			"rpc":    info.FullMethod,
			"req-id": getRequestId(ctx),
			"type":   "unary",
		}
		fields["req"], _ = marshaler.MarshalToString(req.(proto.Message))

		var msg string = "ok"
		if err != nil {
			msg = grpcErrorToJson(err)
		} else if resp != nil {
			fields["resp"], _ = marshaler.MarshalToString(resp.(proto.Message))
		} else {
			msg = "both response data & error is null"
		}

		logger.Withs(fields).Info(msg)

		return
	}

	return f
}

type logServerStream struct {
	grpc.ServerStream

	info   *grpc.StreamServerInfo
	logger slog.Logger

	marshaler *jsonpb.Marshaler
}

func (ss *logServerStream) SendMsg(m interface{}) error {
	s, _ := ss.marshaler.MarshalToString(m.(proto.Message))
	ss.logger.With("data", s).Info("send")

	return ss.ServerStream.SendMsg(m)
}

func (ss *logServerStream) RecvMsg(m interface{}) error {
	err := ss.ServerStream.RecvMsg(m)

	s, _ := ss.marshaler.MarshalToString(m.(proto.Message))
	ss.logger.With("data", s).Info("recv")

	return err
}

func NewLoggingStreamServerInterceptor(cond LogCondition, logger slog.Logger) grpc.StreamServerInterceptor {
	var marshaler = jsonpb.Marshaler{OrigName: true}
	if cond == nil {
		cond = &allwayCondition{}
	}

	f := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !cond.ShouldLogStream(info.FullMethod) {
			return handler(srv, ss)
		}

		fields := slog.Fields{
			"req-id": getRequestId(ss.Context()),
			"type":   "stream",
		}

		currentLogFields := slog.Fields{
			"rpc": info.FullMethod,
		}

		if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
			if agent := md.Get("user-agent"); len(agent) > 0 {
				currentLogFields["agent"] = agent[0]
			}
		}

		baseLog := logger.Withs(fields)

		// log begin stream
		baseLog.Withs(currentLogFields).Info("begin")

		ss = &logServerStream{ss, info, baseLog, &marshaler}
		err := handler(srv, ss)

		var msg = "end"
		if err != nil {
			msg += ": " + grpcErrorToJson(err)
		}
		baseLog.Info(msg)

		return err
	}

	return f
}
