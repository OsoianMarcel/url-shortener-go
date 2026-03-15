package grpcdelivery

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func recoveryUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
		defer func() {
			if recoveredErr := recover(); recoveredErr != nil {
				logger.Error("grpc panic recovered",
					slog.String("method", info.FullMethod),
					slog.Any("error", recoveredErr),
				)
				err = status.Error(codes.Internal, "Internal server error.")
			}
		}()

		return handler(ctx, request)
	}
}

func loggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		response, err := handler(ctx, request)
		duration := time.Since(start)

		grpcCode := status.Code(err)
		logger.Info("grpc request handled",
			slog.String("method", info.FullMethod),
			slog.String("code", grpcCode.String()),
			slog.Duration("duration", duration),
		)

		return response, err
	}
}

func authenticationUnaryInterceptor(apiSecret string, protectedMethods map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, shouldAuthenticate := protectedMethods[info.FullMethod]; !shouldAuthenticate {
			return handler(ctx, request)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "The auth token is missing.")
		}

		authHeader := ""
		for _, value := range md.Get("authorization") {
			trimmedValue := strings.TrimSpace(value)
			if trimmedValue != "" {
				authHeader = trimmedValue
				break
			}
		}

		if authHeader == "" {
			return nil, status.Error(codes.Unauthenticated, "The auth token is missing.")
		}

		authParts := strings.Fields(authHeader)
		if len(authParts) != 2 || !strings.EqualFold(authParts[0], "Bearer") {
			return nil, status.Error(codes.Unauthenticated, "Invalid Authorization header.")
		}

		token := authParts[1]
		if subtle.ConstantTimeCompare([]byte(token), []byte(apiSecret)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "Invalid token.")
		}

		return handler(ctx, request)
	}
}
