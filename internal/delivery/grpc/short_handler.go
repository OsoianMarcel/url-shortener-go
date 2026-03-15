package grpcdelivery

import (
	"context"
	"log/slog"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/grpc/pb"
	"github.com/OsoianMarcel/url-shortener/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ pb.ShortLinkServiceServer = (*shortLinkServer)(nil)

type shortLinkServer struct {
	pb.UnimplementedShortLinkServiceServer
	logger  *slog.Logger
	usecase domain.ShortLinkUsecase
}

func newShortLinkServer(logger *slog.Logger, usecase domain.ShortLinkUsecase) *shortLinkServer {
	return &shortLinkServer{
		logger:  logger,
		usecase: usecase,
	}
}

func (s *shortLinkServer) CreateShortLink(ctx context.Context, request *pb.CreateShortLinkRequest) (*pb.CreateShortLinkResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "Request is required.")
	}

	result, err := s.usecase.Create(ctx, domain.CreateAction{OriginalURL: request.GetUrl()})
	if err != nil {
		if !isHandledDomainError(err) {
			s.logger.Error("GRPC.CreateShortLink", slog.Any("error", err))
		}

		return nil, mapDomainError(err)
	}

	return &pb.CreateShortLinkResponse{
		ShortUrl: result.ShortURL,
		Key:      result.Key,
	}, nil
}

func (s *shortLinkServer) DeleteShortLink(ctx context.Context, request *pb.DeleteShortLinkRequest) (*emptypb.Empty, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "Request is required.")
	}

	err := s.usecase.Delete(ctx, request.GetLinkKey())
	if err != nil {
		if !isHandledDomainError(err) {
			s.logger.Error("GRPC.DeleteShortLink", slog.Any("error", err))
		}

		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *shortLinkServer) ExpandShortLink(ctx context.Context, request *pb.ExpandShortLinkRequest) (*pb.ExpandShortLinkResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "Request is required.")
	}

	entity, err := s.usecase.Expand(ctx, request.GetLinkKey())
	if err != nil {
		if !isHandledDomainError(err) {
			s.logger.Error("GRPC.ExpandShortLink", slog.Any("error", err))
		}

		return nil, mapDomainError(err)
	}

	return &pb.ExpandShortLinkResponse{Url: entity.OriginalURL}, nil
}

func (s *shortLinkServer) GetShortLinkStats(ctx context.Context, request *pb.GetShortLinkStatsRequest) (*pb.GetShortLinkStatsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "Request is required.")
	}

	stats, err := s.usecase.Stats(ctx, request.GetLinkKey())
	if err != nil {
		if !isHandledDomainError(err) {
			s.logger.Error("GRPC.GetShortLinkStats", slog.Any("error", err))
		}

		return nil, mapDomainError(err)
	}

	return &pb.GetShortLinkStatsResponse{
		Hits:      uint64(stats.Hits),
		CreatedAt: timestamppb.New(stats.CreatedAt),
	}, nil
}
