package grpcdelivery

import (
	"log/slog"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/grpc/pb"
	"github.com/OsoianMarcel/url-shortener/internal/domain"
	"google.golang.org/grpc"
)

func NewServer(
	logger *slog.Logger,
	apiSecret string,
	shortLinkUsecase domain.ShortLinkUsecase,
	healthUsecase domain.HealthUsecase,
) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryUnaryInterceptor(logger),
			loggingUnaryInterceptor(logger),
			authenticationUnaryInterceptor(apiSecret, map[string]struct{}{
				pb.ShortLinkService_CreateShortLink_FullMethodName:   {},
				pb.ShortLinkService_DeleteShortLink_FullMethodName:   {},
				pb.ShortLinkService_GetShortLinkStats_FullMethodName: {},
			}),
		),
	)

	pb.RegisterShortLinkServiceServer(server, newShortLinkServer(logger, shortLinkUsecase))
	pb.RegisterHealthServiceServer(server, newHealthServer(healthUsecase))

	return server
}
