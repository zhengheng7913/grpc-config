package client

import (
	"context"
	"fmt"
	"github.com/zhengheng7913/grpc-go-starter/naming/discovery"
	"google.golang.org/grpc"
	"reflect"
)

const GrpcProtocol = "grpc"

var (
	ErrNotGrpcMethod = fmt.Errorf("not a valid grpc method")
	ErrNotGrpcClient = fmt.Errorf("not a valid grpc client")
)

func NewGrpcClient() Client {
	return &GrpcClient{
		options:    nil,
		realClient: nil,
	}
}

type GrpcClient struct {
	options    *Options
	realClient interface{}
}

func (g *GrpcClient) Register(realClient interface{}, opts ...Option) {
	g.options = &Options{}
	for _, opt := range opts {
		opt(g.options)
	}
	d := discovery.Get(g.options.Discovery)
	if d == nil {
		panic(discovery.ErrDiscoveryNotFound)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cons, ok := realClient.(func(cc grpc.ClientConnInterface) interface{})
	if !ok {
		panic(ErrNotGrpcClient)
	}
	target, err := d.Target(fmt.Sprintf("%v://%v", g.options.Discovery, g.options.Name))
	if err != nil {
		panic(fmt.Errorf("get target error: %v", err))
	}
	conn, err := grpc.DialContext(ctx, target)
	if err != nil {
		panic(err)
	}
	g.realClient = cons(conn)

}

func (g *GrpcClient) isGrpcMethod(t reflect.Type) bool {
	return t.NumIn() != 3 || t.NumOut() != 2
}

func (g *GrpcClient) Invoke(context context.Context, method string, req interface{}, opts ...Option) (interface{}, error) {

	handle := reflect.ValueOf(g.realClient).MethodByName(method)
	if !g.isGrpcMethod(handle.Type()) {
		return nil, ErrNotGrpcMethod
	}

	values := handle.Call([]reflect.Value{reflect.ValueOf(context), reflect.ValueOf(req)})
	reply := values[0]
	err := values[1]
	return reply.Interface(), err.Interface().(error)
}