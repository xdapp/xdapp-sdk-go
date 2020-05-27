package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/hprose/hprose-golang/io"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/jhump/protoreflect/desc"
	"github.com/xdapp/xdapp-sdk-go/register"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	AdminIdHeaderKey   = "x-xdapp-admin-id"
	AppIdHeaderKey     = "x-xdapp-app-id"
	ServiceIdHeaderKey = "x-xdapp-service-id"
	RequestIdHeaderKey = "x-xdapp-request-id"
)

type GRPCProxyMiddleware struct {
	descSource    grpcurl.DescriptorSource
	nc            *grpc.ClientConn
	buf           *bytes.Buffer
	rParser       grpcurl.RequestParser
	resolver      jsonpb.AnyResolver
	respFormatter jsonpb.Marshaler
	lastResp      []byte
}

func NewGRPCProxyMiddleware(endpoint string, descFileNames []string, opts ...grpc.DialOption) (*GRPCProxyMiddleware, error) {
	descSource, err := grpcurl.DescriptorSourceFromProtoSets(descFileNames...)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	nc, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	r := grpcurl.AnyResolverFromDescriptorSource(descSource)
	mid := &GRPCProxyMiddleware{
		descSource: descSource,
		nc:         nc,
		buf:        buf,
		resolver:   r,
		respFormatter: jsonpb.Marshaler{
			EnumsAsInts:  true,
			EmitDefaults: true,
			AnyResolver:  r,
		},
	}

	mid.regFunctions()
	return mid, nil
}

func (m *GRPCProxyMiddleware) regFunctions() {
	services, _ := m.descSource.ListServices()
	for _, srvName := range services {
		if m, err := m.descSource.FindSymbol(srvName); err == nil {
			if ms, ok := m.(*desc.ServiceDescriptor); ok {
				for _, sd := range ms.GetMethods() {
					funcName := srvName + "." + sd.GetName()
					register.AddFunction(funcName, func() {})
				}
			}
		}
	}
}

func (m *GRPCProxyMiddleware) OnResolveMethod(descriptor *desc.MethodDescriptor) {
}

func (m *GRPCProxyMiddleware) OnSendHeaders(md metadata.MD) {
}

func (m *GRPCProxyMiddleware) OnReceiveHeaders(md metadata.MD) {
}

func (m *GRPCProxyMiddleware) OnReceiveResponse(message proto.Message) {
	respStr, err := m.respFormatter.MarshalToString(message)
	if err != nil {
		m.lastResp = m.parseRespErr(err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(respStr), &response); err != nil {
		m.lastResp = m.parseRespErr(err)
	}

	writer := io.NewWriter(true)
	writer.WriteByte(io.TagResult)
	writer.Serialize(response)
	writer.WriteByte(io.TagEnd)
	m.lastResp = writer.Bytes()
}

func (m *GRPCProxyMiddleware) parseRespErr(err error) []byte {
	writer := io.NewWriter(true)
	writer.WriteByte(io.TagError)
	writer.WriteString(err.Error())
	writer.WriteByte(io.TagEnd)
	return writer.Bytes()
}

func (m *GRPCProxyMiddleware) OnReceiveTrailers(status *status.Status, md metadata.MD) {
	if status.Code() != codes.OK {
		m.lastResp = m.parseRespErr(status.Err())
	}
}

func (m *GRPCProxyMiddleware) Handler(
	data []byte,
	ctx rpc.Context,
	next rpc.NextFilterHandler) ([]byte, error) {

	method, params, err := m.parseInputData(data)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(method, "sys_") {
		return next(data, ctx)
	}

	header := make([]string, 0)
	header = append(header, fmt.Sprintf("%s: %d", AdminIdHeaderKey, ctx.GetUInt("adminId")))
	header = append(header, fmt.Sprintf("%s: %d", AppIdHeaderKey, ctx.GetUInt("appId")))
	header = append(header, fmt.Sprintf("%s: %d", ServiceIdHeaderKey, ctx.GetUInt("serviceId")))
	header = append(header, fmt.Sprintf("%s: %d", RequestIdHeaderKey, ctx.GetUInt("requestId")))
	return m.requestProxy(context.Background(), m.parseGRPCMethod(method), params, header)
}

func (m *GRPCProxyMiddleware) requestProxy(context context.Context, methodName string, params []interface{}, header []string) ([]byte, error) {
	var data []byte
	var err error

	if len(params) > 0 {
		data, err = json.Marshal(params[0])
		if err != nil {
			return nil, err
		}
	}

	m.buf.Write(data)
	defer m.buf.Reset()
	rf := grpcurl.NewJSONRequestParser(m.buf, m.resolver)
	err = grpcurl.InvokeRPC(context, m.descSource, m.nc, methodName, header, m, rf.Next)
	if err != nil {
		return nil, err
	}

	resp := m.lastResp
	m.lastResp = nil
	return resp, nil
}

func (m *GRPCProxyMiddleware) parseInputData(data []byte) (string, []interface{}, error) {
	var method string
	var params []interface{}

	reader := io.NewReader(data, false)
	reader.JSONCompatible = true
	tag, _ := reader.ReadByte()
	if tag == io.TagCall {
		method = reader.ReadString()
		tag, _ = reader.ReadByte()
		if tag == io.TagList {
			reader.Reset()
			count := reader.ReadCount()
			params = make([]interface{}, count)
			for i := 0; i < count; i++ {
				reader.Unserialize(&params[i])
			}
		}
	}

	return method, params, nil
}

func (m *GRPCProxyMiddleware) parseGRPCMethod(str string) string {
	service := strings.ReplaceAll(str[:strings.LastIndex(str, "_")], "_", ".")
	method := str[strings.LastIndex(str, "_")+1:]
	return service + "/" + method
}
