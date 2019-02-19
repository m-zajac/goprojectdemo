package grpc

import (
	"net"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

// Server can start grpc server handling github most active contributors requests.
type Server struct {
	service ServiceServer
	address string
	l       logrus.FieldLogger
}

// NewServer creates new Server instance.
func NewServer(service ServiceServer, address string, l logrus.FieldLogger) *Server {
	return &Server{
		service: service,
		address: address,
		l:       l,
	}
}

// Run runs the grc server.
// Returns error when failing to open tcp connection.
func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return errors.Wrap(err, "starting tcp listener")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	srv := grpc.NewServer()
	RegisterServiceServer(srv, s.service)

	go func() {
		s.l.Infof("starting grpc server, listening on %s", s.address)
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			s.l.Errorf("grpc server returned error: %v", err)
		}
	}()

	<-stop
	srv.GracefulStop()
	s.l.Info("grpc server shut down")

	return nil
}
