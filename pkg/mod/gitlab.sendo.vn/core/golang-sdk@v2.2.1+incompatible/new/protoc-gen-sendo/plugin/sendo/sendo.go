package sendo

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"gitlab.sendo.vn/core/golang-sdk/new/protoc-gen-sendo/generator"
)

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	contextPkgPath = "context"
	sendoPkgPath   = "gitlab.sendo.vn/core/golang-sdk/new"
)

func init() {
	generator.RegisterPlugin(new(sendo))
}

// sendo is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for go-sendo support.
type sendo struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "sendo".
func (g *sendo) Name() string {
	return "sendo"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	contextPkg string
	sendoSDK   string
)

// Init initializes the plugin.
func (g *sendo) Init(gen *generator.Generator) {
	g.gen = gen
	contextPkg = generator.RegisterUniquePackageName("context", nil)
	sendoSDK = generator.RegisterUniquePackageName("sendo", nil)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *sendo) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *sendo) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *sendo) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *sendo) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("// Reference imports to suppress errors if they are not otherwise used.")
	g.P("var _ ", contextPkg, ".Context")
	g.P()

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *sendo) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.P(contextPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, contextPkgPath)))
	g.P(sendoSDK, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, sendoPkgPath)))
	g.P(strconv.Quote("google.golang.org/grpc"))
	g.P(")")
	g.P()
}

// reservedClientName records whether a client name is reserved on the client side.
var reservedClientName = map[string]bool{
	// TODO: do we need any in go-sendo?
}

func unexport(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// generateService generates all the code for the named service.
func (g *sendo) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	path := fmt.Sprintf("6,%d", index) // 6 means service.

	origServName := service.GetName()
	servName := generator.CamelCase(origServName)
	servAlias := servName + "Service"

	// strip suffix
	if strings.HasSuffix(servAlias, "ServiceService") {
		servAlias = strings.TrimSuffix(servAlias, "Service")
	}

	g.P()
	g.P("// Client API for ", servName, " service")
	g.P()

	sdServer := "SD" + servName + "Server"
	// SD Server interface.
	g.P("type ", servAlias, " interface {")
	for i, method := range service.Method {
		g.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		g.P(g.generateClientSignature(servName, method))
	}
	g.P("}")
	g.P()

	// SD Server structure.
	g.P("type ", unexport(sdServer), " struct {")
	g.P("ctx ", sendoSDK, ".ServiceContext")
	g.P("sv ", servAlias)
	g.P("}")
	g.P()

	// NewClient factory.
	g.P("func New", sdServer, " (ctx ", sendoSDK, ".ServiceContext", ", sv ", servAlias, ") *", unexport(sdServer), "{")
	g.P("return &", unexport(sdServer), "{")
	g.P("sv: sv,")
	g.P("ctx: ctx,")
	g.P("}")
	g.P("}")
	g.P()

	for _, method := range service.Method {
		g.generateServerMethod(servName, method)
	}

	g.P("func Add", servName, "Handler(service sendo.Service, sv ", servName, "){")
	g.P("service.Server().AddHandler(func(gRPC *grpc.Server) {")
	g.P("Register", servName, "Server(gRPC, New", sdServer, "(service, sv))")
	g.P("})")
	g.P("}")
	g.P()
}

// generateClientSignature returns the client-side signature for a method.
func (g *sendo) generateClientSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}
	reqArg := ", in *" + g.typeName(method.GetInputType())
	if method.GetClientStreaming() {
		reqArg = ""
	}
	respName := "*" + g.typeName(method.GetOutputType())

	if method.GetServerStreaming() || method.GetClientStreaming() {
		respName = servName + "_" + generator.CamelCase(methName) + "Server"
		return fmt.Sprintf("%s(sdCtx sendo.ServiceContext%s, streamServer %s) error", methName, reqArg, respName)
	}

	return fmt.Sprintf("%s(sdCtx sendo.ServiceContext, ctx %s.Context%s) (%s, error)", methName, contextPkg, reqArg, respName)
}

func (g *sendo) generateServerMethod(servName string, method *pb.MethodDescriptorProto) {
	methName := generator.CamelCase(method.GetName())
	sdServer := "SD" + servName + "Server"
	inType := g.typeName(method.GetInputType())
	outType := g.typeName(method.GetOutputType())

	respName := servName + "_" + generator.CamelCase(methName) + "Server"
	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		g.P("func (h *", unexport(sdServer), ") ", methName, "(ctx ", contextPkg, ".Context, in *", inType, ") (*", outType, ", error) {")
		g.P("return h.sv", ".", methName, "(h.ctx, ctx, in)")
		g.P("}")
		g.P()
	} else if method.GetServerStreaming() && method.GetClientStreaming() {
		g.P("func (h *", unexport(sdServer), ") ", methName, "(streamSV ", respName, ")  error {")
		g.P("return h.sv", ".", methName, "(h.ctx, streamSV)")
		g.P("}")
		g.P()
	} else {
		// TODO: Stream method, this is temporary to finish implement
		respName := servName + "_" + generator.CamelCase(methName) + "Server"
		g.P("func (h *", unexport(sdServer), ") ", methName, "(in *", inType, ", streamSV ", respName, ")  error {")
		g.P("return h.sv", ".", methName, "(h.ctx, in, streamSV)")
		g.P("}")
		g.P()
	}

}
