package grpcdelivery

import (
	"context"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/grpc/pb"
	"github.com/OsoianMarcel/url-shortener/internal/domain"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ pb.HealthServiceServer = (*healthServer)(nil)

type healthServer struct {
	pb.UnimplementedHealthServiceServer
	usecase domain.HealthUsecase
}

func newHealthServer(usecase domain.HealthUsecase) *healthServer {
	return &healthServer{usecase: usecase}
}

func (s *healthServer) CheckHealth(ctx context.Context, _ *pb.CheckHealthRequest) (*pb.CheckHealthResponse, error) {
	checkResult := s.usecase.CheckHealth(ctx)
	services := make([]*pb.ServiceHealth, 0, len(checkResult.Services))

	for _, service := range checkResult.Services {
		services = append(services, &pb.ServiceHealth{
			Name:          service.Name,
			Healthy:       service.Healthy,
			Error:         service.Error,
			CheckDuration: durationpb.New(service.CheckDuration),
		})
	}

	return &pb.CheckHealthResponse{
		AllHealthy: checkResult.AllHealthy,
		Services:   services,
		ServerTime: timestamppb.Now(),
	}, nil
}
