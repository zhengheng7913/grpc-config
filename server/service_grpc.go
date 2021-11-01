package server

import (
	"errors"
	"fmt"
	"github.com/zhengheng7913/grpc-config/config"
	"google.golang.org/grpc"
	"log"
	"net"
)

func newServiceRegisterAdapter(srv Service) grpc.ServiceRegistrar {
	return &ServiceRegisterAdapter{
		service: srv,
	}
}

type GrpcService struct {
	server *grpc.Server
	cfg    *config.ServiceConfig
	opt    *Options
}

func (g *GrpcService) Register(serviceDesc interface{}, serviceImpl interface{}) {
	desc, ok := serviceDesc.(*grpc.ServiceDesc)
	if !ok {
		fmt.Println(errors.New("service desc type invalid"))
		return
	}
	opts := g.opt.ServiceOptions.(*GrpcOptions).Opts
	g.server = grpc.NewServer(opts...)
	g.server.RegisterService(desc, serviceImpl)
}

func (g *GrpcService) Serve() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", g.cfg.Port))
	if err != nil {
		return fmt.Errorf("Failed to listen: %v ", err)
	}
	go func() {
		err := g.server.Serve(lis)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	err = g.opt.Registry.Register(g.cfg.Name)
	if err != nil {
		return err
	}
	return nil
}

func (g *GrpcService) Close(c chan struct{}) error {
	g.server.Stop()
	return nil
}

type ServiceRegisterAdapter struct {
	service Service
}

func (s *ServiceRegisterAdapter) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.service.Register(desc, impl)
}

type GrpcOptions struct {
	Opts []grpc.ServerOption
}

func (o GrpcOptions) ProtocolName() string {
	return ProtocolNameGrpc
}

func (o *GrpcOptions) Apply(inters ...interface{}) {
	gOpts, ok := assertGrpcOptions(inters...)
	if !ok {
		panic("unknown service type")
	}
	o.Opts = append(o.Opts, gOpts...)
}

func assertGrpcOptions(inters ...interface{}) ([]grpc.ServerOption, bool) {
	opts := make([]grpc.ServerOption, len(inters))
	for _, inter := range inters {
		opt, ok := inter.(grpc.ServerOption)
		if !ok {
			return nil, false
		}
		opts = append(opts, opt)
	}
	return opts, true
}

func dessertGrpcOptions(opts ...grpc.ServerOption) []interface{} {
	inters := make([]interface{}, len(opts))
	for _, opt := range opts {
		inter := opt.(interface{})
		inters = append(inters, inter)
	}
	return inters
}