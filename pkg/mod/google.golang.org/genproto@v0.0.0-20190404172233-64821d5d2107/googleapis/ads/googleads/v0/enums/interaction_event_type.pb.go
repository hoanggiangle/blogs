// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v0/enums/interaction_event_type.proto

package enums // import "google.golang.org/genproto/googleapis/ads/googleads/v0/enums"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Enum describing possible types of payable and free interactions.
type InteractionEventTypeEnum_InteractionEventType int32

const (
	// Not specified.
	InteractionEventTypeEnum_UNSPECIFIED InteractionEventTypeEnum_InteractionEventType = 0
	// Used for return value only. Represents value unknown in this version.
	InteractionEventTypeEnum_UNKNOWN InteractionEventTypeEnum_InteractionEventType = 1
	// Click to site. In most cases, this interaction navigates to an external
	// location, usually the advertiser's landing page. This is also the default
	// InteractionEventType for click events.
	InteractionEventTypeEnum_CLICK InteractionEventTypeEnum_InteractionEventType = 2
	// The user's expressed intent to engage with the ad in-place.
	InteractionEventTypeEnum_ENGAGEMENT InteractionEventTypeEnum_InteractionEventType = 3
	// User viewed a video ad.
	InteractionEventTypeEnum_VIDEO_VIEW InteractionEventTypeEnum_InteractionEventType = 4
	// The default InteractionEventType for ad conversion events.
	// This is used when an ad conversion row does NOT indicate
	// that the free interactions (i.e., the ad conversions)
	// should be 'promoted' and reported as part of the core metrics.
	// These are simply other (ad) conversions.
	InteractionEventTypeEnum_NONE InteractionEventTypeEnum_InteractionEventType = 5
)

var InteractionEventTypeEnum_InteractionEventType_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "UNKNOWN",
	2: "CLICK",
	3: "ENGAGEMENT",
	4: "VIDEO_VIEW",
	5: "NONE",
}
var InteractionEventTypeEnum_InteractionEventType_value = map[string]int32{
	"UNSPECIFIED": 0,
	"UNKNOWN":     1,
	"CLICK":       2,
	"ENGAGEMENT":  3,
	"VIDEO_VIEW":  4,
	"NONE":        5,
}

func (x InteractionEventTypeEnum_InteractionEventType) String() string {
	return proto.EnumName(InteractionEventTypeEnum_InteractionEventType_name, int32(x))
}
func (InteractionEventTypeEnum_InteractionEventType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_interaction_event_type_74dff4de963ff5a2, []int{0, 0}
}

// Container for enum describing types of payable and free interactions.
type InteractionEventTypeEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InteractionEventTypeEnum) Reset()         { *m = InteractionEventTypeEnum{} }
func (m *InteractionEventTypeEnum) String() string { return proto.CompactTextString(m) }
func (*InteractionEventTypeEnum) ProtoMessage()    {}
func (*InteractionEventTypeEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_interaction_event_type_74dff4de963ff5a2, []int{0}
}
func (m *InteractionEventTypeEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InteractionEventTypeEnum.Unmarshal(m, b)
}
func (m *InteractionEventTypeEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InteractionEventTypeEnum.Marshal(b, m, deterministic)
}
func (dst *InteractionEventTypeEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InteractionEventTypeEnum.Merge(dst, src)
}
func (m *InteractionEventTypeEnum) XXX_Size() int {
	return xxx_messageInfo_InteractionEventTypeEnum.Size(m)
}
func (m *InteractionEventTypeEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_InteractionEventTypeEnum.DiscardUnknown(m)
}

var xxx_messageInfo_InteractionEventTypeEnum proto.InternalMessageInfo

func init() {
	proto.RegisterType((*InteractionEventTypeEnum)(nil), "google.ads.googleads.v0.enums.InteractionEventTypeEnum")
	proto.RegisterEnum("google.ads.googleads.v0.enums.InteractionEventTypeEnum_InteractionEventType", InteractionEventTypeEnum_InteractionEventType_name, InteractionEventTypeEnum_InteractionEventType_value)
}

func init() {
	proto.RegisterFile("google/ads/googleads/v0/enums/interaction_event_type.proto", fileDescriptor_interaction_event_type_74dff4de963ff5a2)
}

var fileDescriptor_interaction_event_type_74dff4de963ff5a2 = []byte{
	// 321 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x90, 0xd1, 0x4e, 0xf2, 0x30,
	0x1c, 0xc5, 0xbf, 0x0d, 0xf8, 0xd4, 0x92, 0xe8, 0xb2, 0x78, 0xa1, 0x17, 0x5c, 0xc0, 0x03, 0x74,
	0x4b, 0xbc, 0xab, 0x57, 0x05, 0xea, 0xd2, 0xa0, 0x85, 0x44, 0x18, 0x89, 0x59, 0x42, 0x26, 0x6b,
	0x9a, 0x25, 0xd0, 0x2e, 0x74, 0x90, 0xf0, 0x00, 0xbe, 0x88, 0x97, 0x3e, 0x8a, 0x8f, 0xe2, 0x85,
	0xcf, 0x60, 0xda, 0xc9, 0xbc, 0x41, 0x6f, 0x9a, 0xd3, 0x9e, 0xff, 0xaf, 0x39, 0xff, 0x03, 0x90,
	0x50, 0x4a, 0xac, 0x78, 0x90, 0x66, 0x3a, 0xa8, 0xa4, 0x51, 0xbb, 0x30, 0xe0, 0x72, 0xbb, 0xd6,
	0x41, 0x2e, 0x4b, 0xbe, 0x49, 0x97, 0x65, 0xae, 0xe4, 0x82, 0xef, 0xb8, 0x2c, 0x17, 0xe5, 0xbe,
	0xe0, 0xb0, 0xd8, 0xa8, 0x52, 0xf9, 0x9d, 0x0a, 0x80, 0x69, 0xa6, 0x61, 0xcd, 0xc2, 0x5d, 0x08,
	0x2d, 0xdb, 0x7b, 0x71, 0xc0, 0x15, 0xfd, 0xe1, 0x89, 0xc1, 0xa7, 0xfb, 0x82, 0x13, 0xb9, 0x5d,
	0xf7, 0x72, 0x70, 0x79, 0xcc, 0xf3, 0x2f, 0x40, 0x7b, 0xc6, 0x1e, 0x27, 0x64, 0x40, 0xef, 0x28,
	0x19, 0x7a, 0xff, 0xfc, 0x36, 0x38, 0x99, 0xb1, 0x11, 0x1b, 0xcf, 0x99, 0xe7, 0xf8, 0x67, 0xa0,
	0x35, 0xb8, 0xa7, 0x83, 0x91, 0xe7, 0xfa, 0xe7, 0x00, 0x10, 0x16, 0xe1, 0x88, 0x3c, 0x10, 0x36,
	0xf5, 0x1a, 0xe6, 0x1e, 0xd3, 0x21, 0x19, 0x2f, 0x62, 0x4a, 0xe6, 0x5e, 0xd3, 0x3f, 0x05, 0x4d,
	0x36, 0x66, 0xc4, 0x6b, 0xf5, 0x3f, 0x1d, 0xd0, 0x5d, 0xaa, 0x35, 0xfc, 0x33, 0x6d, 0xff, 0xfa,
	0x58, 0x9c, 0x89, 0xd9, 0x73, 0xe2, 0x3c, 0xf5, 0xbf, 0x59, 0xa1, 0x56, 0xa9, 0x14, 0x50, 0x6d,
	0x44, 0x20, 0xb8, 0xb4, 0x2d, 0x1c, 0x5a, 0x2b, 0x72, 0xfd, 0x4b, 0x89, 0xb7, 0xf6, 0x7c, 0x75,
	0x1b, 0x11, 0xc6, 0x6f, 0x6e, 0x27, 0xaa, 0xbe, 0xc2, 0x99, 0x86, 0x95, 0x34, 0x2a, 0x0e, 0xa1,
	0xa9, 0x45, 0xbf, 0x1f, 0xfc, 0x04, 0x67, 0x3a, 0xa9, 0xfd, 0x24, 0x0e, 0x13, 0xeb, 0x7f, 0xb8,
	0xdd, 0xea, 0x11, 0x21, 0x9c, 0x69, 0x84, 0xea, 0x09, 0x84, 0xe2, 0x10, 0x21, 0x3b, 0xf3, 0xfc,
	0xdf, 0x06, 0xbb, 0xf9, 0x0a, 0x00, 0x00, 0xff, 0xff, 0x5a, 0xa2, 0x76, 0xb7, 0xdc, 0x01, 0x00,
	0x00,
}
