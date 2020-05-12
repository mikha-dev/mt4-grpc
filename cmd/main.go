package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"mt4grpc/api_pb"
	"mt4grpc/common"
	"mt4grpc/service"
	"mtconfig"
	"mtdealer"
	"mtlog"
	"net"
	"net/http"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sherifabdlnaby/configuro"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Config struct {
	Server struct {
		RestAddr string
		GrpcAddr string
		PromAddr string
		WsAddr   string
		Mode     string
	}

	ManagerLogger mtconfig.Common

	Dealers map[string]mtdealer.Config
}

var conf *Config
var log *zap.Logger

func initLog() error {

	// loading mt4 api manager logger
	l, err := mtlog.NewLogger(conf.ManagerLogger.LogPath, conf.ManagerLogger.LogLevel)
	if err != nil {
		return err
	}

	mtlog.SetDefault(l)
	mtlog.Info("log path: \"%s\" with level \"%s\"", conf.ManagerLogger.LogPath, conf.ManagerLogger.LogLevel)
	// -- loading mt4 api manager logger

	// loading zap logger
	var cfg zap.Config
	f, _ := ioutil.ReadFile(".\\logger.json")
	if err := json.Unmarshal(f, &cfg); err != nil {
		panic(err)
	}
	if log, err = cfg.Build(); err != nil {
		panic(err)
	}
	// -- loading zap logger

	return nil
}

func run() error {

	configLoader, err := configuro.NewConfig()
	if err != nil {
		return err
	}

	conf = &Config{}

	if err := configLoader.Load(conf); err != nil {
		return err
	}

	if err := initLog(); err != nil {
		panic(err)
	}

	log.Info("loader servers",
		zap.String("rest", conf.Server.RestAddr),
		zap.String("grpc", conf.Server.GrpcAddr),
		zap.String("prometheus", conf.Server.PromAddr),
		zap.String("ws", conf.Server.WsAddr),
	)

	for k, v := range conf.Dealers {
		log.Info(
			"loaded manager config",
			zap.String("token", k),
			zap.String("server", v.ServerAddr),
			zap.Int("accont", v.Account),
		)
	}
	loader := common.NewDealerLoader(conf.Dealers)

	defer func() {
		loader.Stop()
	}()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	lis, err := net.Listen("tcp", conf.Server.GrpcAddr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	api_pb.RegisterUserServiceServer(grpcServer, service.NewUserService(loader, log))
	grpc_prometheus.Register(grpcServer)

	var group errgroup.Group

	group.Go(func() error {
		return grpcServer.Serve(lis)
	})

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	runtime.SetHTTPBodyMarshaler(mux)
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50000000)),
	}

	hmux := http.NewServeMux()
	hmux.HandleFunc("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "api.swagger.json")
	})

	hmux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "doc.html")
	})

	group.Go(func() error {
		return http.ListenAndServe(conf.Server.RestAddr, hmux)
	})

	group.Go(func() error {
		return api_pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, conf.Server.GrpcAddr, opts)
	})
	group.Go(func() error {
		return http.ListenAndServe(conf.Server.WsAddr, wsproxy.WebsocketProxy(mux))
	})
	group.Go(func() error {
		return http.ListenAndServe(conf.Server.PromAddr, promhttp.Handler())
	})

	return group.Wait()
}

func main() {
	flag.Parse()

	defer func() {
		if err := log.Sync(); err != nil {
			panic(err)
		}
	}()

	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
