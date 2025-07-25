// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: screens/v1/screens.proto

package screensv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/DaanV2/mechanus/server/pkg/grpc/gen/screens/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// ScreensServiceName is the fully-qualified name of the ScreensService service.
	ScreensServiceName = "screens.v1.ScreensService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ScreensServiceListenUpdateProcedure is the fully-qualified name of the ScreensService's
	// ListenUpdate RPC.
	ScreensServiceListenUpdateProcedure = "/screens.v1.ScreensService/ListenUpdate"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	screensServiceServiceDescriptor            = v1.File_screens_v1_screens_proto.Services().ByName("ScreensService")
	screensServiceListenUpdateMethodDescriptor = screensServiceServiceDescriptor.Methods().ByName("ListenUpdate")
)

// ScreensServiceClient is a client for the screens.v1.ScreensService service.
type ScreensServiceClient interface {
	ListenUpdate(context.Context, *connect.Request[v1.ScreenListenRequest]) (*connect.ServerStreamForClient[v1.ScreenUpdate], error)
}

// NewScreensServiceClient constructs a client for the screens.v1.ScreensService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewScreensServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ScreensServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &screensServiceClient{
		listenUpdate: connect.NewClient[v1.ScreenListenRequest, v1.ScreenUpdate](
			httpClient,
			baseURL+ScreensServiceListenUpdateProcedure,
			connect.WithSchema(screensServiceListenUpdateMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// screensServiceClient implements ScreensServiceClient.
type screensServiceClient struct {
	listenUpdate *connect.Client[v1.ScreenListenRequest, v1.ScreenUpdate]
}

// ListenUpdate calls screens.v1.ScreensService.ListenUpdate.
func (c *screensServiceClient) ListenUpdate(ctx context.Context, req *connect.Request[v1.ScreenListenRequest]) (*connect.ServerStreamForClient[v1.ScreenUpdate], error) {
	return c.listenUpdate.CallServerStream(ctx, req)
}

// ScreensServiceHandler is an implementation of the screens.v1.ScreensService service.
type ScreensServiceHandler interface {
	ListenUpdate(context.Context, *connect.Request[v1.ScreenListenRequest], *connect.ServerStream[v1.ScreenUpdate]) error
}

// NewScreensServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewScreensServiceHandler(svc ScreensServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	screensServiceListenUpdateHandler := connect.NewServerStreamHandler(
		ScreensServiceListenUpdateProcedure,
		svc.ListenUpdate,
		connect.WithSchema(screensServiceListenUpdateMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/screens.v1.ScreensService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ScreensServiceListenUpdateProcedure:
			screensServiceListenUpdateHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedScreensServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedScreensServiceHandler struct{}

func (UnimplementedScreensServiceHandler) ListenUpdate(context.Context, *connect.Request[v1.ScreenListenRequest], *connect.ServerStream[v1.ScreenUpdate]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("screens.v1.ScreensService.ListenUpdate is not implemented"))
}
