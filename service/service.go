// Package service handles setting up a gRPC service with
// all the bells and whistles required to operate it.
package service

import (
	"net"
	"net/http"
	_ "net/http/pprof" // register pprof in the http.DefaultServeMux
	"strconv"
	"time"

	"github.com/aybabtme/std/flag"
	"github.com/aybabtme/std/log"
	"github.com/aybabtme/std/metric"
	"github.com/aybabtme/std/service/discover"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// AddrsFunc is required by a service so that it know what addresses
// to listen on and where to discover services. It is evaluated only
// once the flags have been parsed.
type AddrsFunc func() (string, string, string)

// Addrs returns the addresses a service will use to provide
// rpc (gRPC) and status (HTTP), and the address where to reach
// the discovery service.
func Addrs(rpcAddr, statusAddr, discoveryAddr *string) AddrsFunc {
	return func() (string, string, string) {
		if rpcAddr == nil {
			panic("rpc addr is nil, they need to be set before the service starts")
		}
		if statusAddr == nil {
			panic("status addr is nil, they need to be set before the service starts")
		}
		if discoveryAddr == nil {
			panic("discovery addr is nil, they need to be set before the service starts")
		}
		return *rpcAddr, *statusAddr, *discoveryAddr
	}
}

// Start prepares and runs a service until it fails, somehow.
func Start(
	appname string,
	setFlag func(fs flag.FlagSet),
	addr AddrsFunc,
	registerFn func(log.Log, metric.Node, discover.Dialer, *grpc.Server),
) {
	ll := log.KV("appname", appname)
	ll.Info("preparing")

	sys, promhdl := metric.Prometheus()
	sys = sys.Lbl("appname", appname)

	ll.Info("parsing flags")
	err := flag.ParseSet(appname, setFlag)
	if err != nil {
		ll.Err(err).Fatal("invalid flag or environment")
	}

	rpcAddr, statusAddr, discoveryAddr := addr()

	rpcl, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		ll.Err(err).Fatal("can't bind grpc address")
	}

	statusl, err := net.Listen("tcp", statusAddr)
	if err != nil {
		ll.Err(err).Fatal("can't bind status address")
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(unaryIceptor(sys)),
		grpc.StreamInterceptor(streamIceptor(sys)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	ll.Info("registering with discovery")

	dialer, err := discover.Register(ctx, discoveryAddr, discover.ServiceDesc{
		Name:       appname,
		RPCAddr:    rpcAddr,
		StatusAddr: statusAddr,
	})
	if err != nil {
		ll.KV("discover.addr", discoveryAddr).Err(err).Fatal("can't register with discovery service")
	}

	ll.Info("registering service")
	registerFn(ll, sys, dialer, s)

	ll.KV("addr", statusl.Addr()).Info("starting status server")
	go func(ctx context.Context) {
		defer cancel()

		mux := http.DefaultServeMux
		mux.Handle("/metrics", promhdl)

		err = (&http.Server{Handler: mux}).Serve(statusl)
		ll.Err(err).Error("http server quit unexpectedly")
	}(ctx)

	ll.KV("addr", rpcl.Addr()).Info("starting rpc server")
	go func(ctx context.Context) {
		defer cancel()

		err = s.Serve(rpcl)
		ll.Err(err).Error("rpc server quit unexpectedly")
	}(ctx)

	<-ctx.Done()
	if err := rpcl.Close(); err != nil {
		ll.Err(err).Warn("failed to close rpc listener")
	}
	if err := statusl.Close(); err != nil {
		ll.Err(err).Warn("failed to close status listener")
	}
}

func unaryIceptor(sys metric.Node) grpc.UnaryServerInterceptor {
	var (
		requestTotal = sys.Counter(
			"unary_request_total",
			"Number of unary requests received",
			"method",
		)
		responseTotal = sys.Counter(
			"unary_response_total",
			"Number of unary responses sent",
			"method", "code",
		)
		responseDuration = sys.Sampler(
			"unary_response_duration_seconds",
			"Seconds elapsed to produce a unary response",
			0.008, 30,
			"method", "code",
		)
	)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		method := info.FullMethod

		requestTotal(1, method)

		start := time.Now()
		resp, err = handler(ctx, req)

		code := grpc.Code(err).String()
		responseDuration(time.Since(start).Seconds(), method, code)
		responseTotal(1, method, code)

		return resp, err
	}
}

func streamIceptor(sys metric.Node) grpc.StreamServerInterceptor {
	var (
		requestTotal = sys.Counter(
			"stream_request_total",
			"Number of stream requests received",
			"method", "client_stream", "server_stream",
		)
		responseTotal = sys.Counter(
			"stream_response_total",
			"Number of stream responses sent",
			"method", "client_stream", "server_stream", "code",
		)
		responseDuration = sys.Sampler(
			"stream_response_duration_seconds",
			"Seconds elapsed to produce a stream response",
			0.008, 30,
			"method", "client_stream", "server_stream", "code",
		)
	)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		var (
			method       = info.FullMethod
			clientStream = strconv.FormatBool(info.IsClientStream)
			serverStream = strconv.FormatBool(info.IsServerStream)
		)

		requestTotal(1, method, clientStream, serverStream)

		start := time.Now()
		err := handler(srv, ss)

		code := grpc.Code(err).String()
		responseDuration(time.Since(start).Seconds(), method, clientStream, serverStream, code)
		responseTotal(1, method, clientStream, serverStream, code)

		return err
	}
}
