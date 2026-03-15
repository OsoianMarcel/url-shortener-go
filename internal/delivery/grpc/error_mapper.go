package grpcdelivery

import (
	"errors"

	"github.com/OsoianMarcel/url-shortener/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidURL):
		return status.Error(codes.InvalidArgument, "Invalid URL.")
	case errors.Is(err, domain.ErrShortLinkNotFound):
		return status.Error(codes.NotFound, "Link not found.")
	default:
		return status.Error(codes.Internal, "Internal server error.")
	}
}

func isHandledDomainError(err error) bool {
	return errors.Is(err, domain.ErrInvalidURL) || errors.Is(err, domain.ErrShortLinkNotFound)
}
