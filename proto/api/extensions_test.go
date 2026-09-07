package api

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestExtensionWireCompatibility(t *testing.T) {
	tests := []struct {
		name       string
		extension  protoreflect.ExtensionType
		fullName   protoreflect.FullName
		number     protoreflect.FieldNumber
		extendee   protoreflect.FullName
		newMessage func() proto.Message
	}{
		{
			name:       "http",
			extension:  E_Http,
			fullName:   "pzero.api.http",
			number:     10000,
			extendee:   "google.protobuf.MethodOptions",
			newMessage: func() proto.Message { return &descriptorpb.MethodOptions{} },
		},
		{
			name:       "http_group",
			extension:  E_HttpGroup,
			fullName:   "pzero.api.http_group",
			number:     10001,
			extendee:   "google.protobuf.ServiceOptions",
			newMessage: func() proto.Message { return &descriptorpb.ServiceOptions{} },
		},
		{
			name:       "zrpc",
			extension:  E_Zrpc,
			fullName:   "pzero.api.zrpc",
			number:     10002,
			extendee:   "google.protobuf.MethodOptions",
			newMessage: func() proto.Message { return &descriptorpb.MethodOptions{} },
		},
		{
			name:       "zrpc_group",
			extension:  E_ZrpcGroup,
			fullName:   "pzero.api.zrpc_group",
			number:     10003,
			extendee:   "google.protobuf.ServiceOptions",
			newMessage: func() proto.Message { return &descriptorpb.ServiceOptions{} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descriptor := tt.extension.TypeDescriptor()
			require.Equal(t, tt.fullName, descriptor.FullName())
			require.Equal(t, tt.number, descriptor.Number())
			require.Equal(t, tt.extendee, descriptor.ContainingMessage().FullName())

			options := tt.newMessage()
			switch tt.extension {
			case E_Http, E_HttpGroup:
				proto.SetExtension(options, tt.extension, &HttpRule{Middleware: "auth"})
			case E_Zrpc, E_ZrpcGroup:
				proto.SetExtension(options, tt.extension, &ZrpcRule{Middleware: "auth"})
			}

			wire, err := proto.Marshal(options)
			require.NoError(t, err)
			decoded := tt.newMessage()
			require.NoError(t, proto.Unmarshal(wire, decoded))
			require.True(t, proto.HasExtension(decoded, tt.extension))
		})
	}
}
