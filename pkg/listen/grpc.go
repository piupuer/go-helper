package listen

import (
	"fmt"
	"github.com/piupuer/go-helper/pkg/rpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func Grpc(options ...func(*GrpcOptions)) {
	ops := getGrpcOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}

	if ops.logger != nil {
		ops.serverOps = append(ops.serverOps, rpc.WithGrpcServerLogger(ops.logger))
	}
	if ops.ctx != nil {
		ops.serverOps = append(ops.serverOps, rpc.WithGrpcServerCtx(ops.ctx))
	}
	// create grpc server
	srv := rpc.NewGrpcServer(ops.serverOps...)
	// register servers
	if ops.register != nil {
		ops.register(srv)
	}

	host := ops.host
	port := ops.port
	addr := fmt.Sprintf("%s:%d", host, port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		ops.logger.Error("[%s][grpc server]failed to listen: %v", ops.proName, err)
		return
	}

	go func() {
		if err = srv.Serve(lis); err != nil {
			ops.logger.Error("[%s][grpc server]failed to serve: %v", ops.proName, err)
		}
	}()

	ops.logger.Info("[%s][grpc server]running at %s:%d", ops.proName, host, port)

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ops.logger.Info("[%s][grpc server]shutting down...", ops.proName)

	srv.GracefulStop()
	ops.logger.Info("[%s][grpc server]exiting", ops.proName)
}
