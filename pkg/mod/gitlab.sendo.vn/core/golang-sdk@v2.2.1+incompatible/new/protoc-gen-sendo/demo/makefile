GO_PROTOC   ?= gogoslick
ifneq (go,$(GO_PROTOC))
GOGO_FLAGS   = Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,
GOGO_FLAGS  := $(GOGO_FLAGS)Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/types,
GOGO_FLAGS  := $(GOGO_FLAGS)Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,
GOGO_FLAGS  := $(GOGO_FLAGS)Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,
GOGO_FLAGS  := $(GOGO_FLAGS)Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,
GOGO_FLAGS  := $(GOGO_FLAGS)Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,
GOGO_FLAGS  := $(GOGO_FLAGS)Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,
GOGO_FLAGS  := $(GOGO_FLAGS)Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,
GOGO_FLAGS  := $(GOGO_FLAGS)Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,
endif

gen:
	protoc -I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. --sendo_out=. test.proto