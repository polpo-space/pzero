package genrpc

import (
	"fmt"
	"testing"

	"github.com/jhump/protoreflect/desc/protoparse"
	pzeroapi "github.com/polpo-space/pzero/proto/api"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestMiddlewareExtensionCompatibility(t *testing.T) {
	for _, namespace := range []string{"pzero", "jzero"} {
		t.Run(namespace, func(t *testing.T) {
			files := map[string]string{
				"service.proto": fmt.Sprintf(`
syntax = "proto3";
package example;
import %q;
import %q;
service Example {
  option (%s.api.http_group) = { middleware: "auth" };
  rpc Call(Request) returns (Response) {
    option (%s.api.zrpc) = { middleware: "trace" };
  }
}
message Request {}
message Response {}
`, namespace+"/api/http.proto", namespace+"/api/zrpc.proto", namespace, namespace),
				namespace + "/api/http.proto": middlewareProto(namespace, "http"),
				namespace + "/api/zrpc.proto": middlewareProto(namespace, "zrpc"),
			}

			parser := protoparse.Parser{Accessor: protoparse.FileContentsFromMap(files)}
			descriptors, err := parser.ParseFiles("service.proto")
			require.NoError(t, err)

			service := descriptors[0].AsFileDescriptorProto().GetService()[0]
			httpRule, ok := proto.GetExtension(service.GetOptions(), pzeroapi.E_HttpGroup).(*pzeroapi.HttpRule)
			require.True(t, ok)
			require.Equal(t, "auth", httpRule.GetMiddleware())

			zrpcRule, ok := proto.GetExtension(service.GetMethod()[0].GetOptions(), pzeroapi.E_Zrpc).(*pzeroapi.ZrpcRule)
			require.True(t, ok)
			require.Equal(t, "trace", zrpcRule.GetMiddleware())
		})
	}
}

func middlewareProto(namespace, kind string) string {
	if kind == "http" {
		return fmt.Sprintf(`
syntax = "proto3";
package %s.api;
import "google/protobuf/descriptor.proto";
message HttpRule { string middleware = 1; }
extend google.protobuf.MethodOptions { HttpRule http = 10000; }
extend google.protobuf.ServiceOptions { HttpRule http_group = 10001; }
`, namespace)
	}
	return fmt.Sprintf(`
syntax = "proto3";
package %s.api;
import "google/protobuf/descriptor.proto";
message ZrpcRule { string middleware = 1; }
extend google.protobuf.MethodOptions { ZrpcRule zrpc = 10002; }
extend google.protobuf.ServiceOptions { ZrpcRule zrpc_group = 10003; }
`, namespace)
}
