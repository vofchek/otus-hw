package internalgrpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func LogInterceptor(logg Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		result, err := handler(ctx, req)

		meta, _ := metadata.FromIncomingContext(ctx)

		logg.Info(
			fmt.Sprintf(`%s [%s] %s %s %s %d %d %s`,
				"-",                             // IP клиента
				time.Now().Format(time.RFC3339), // дата и время запроса
				"-",                             // метод
				info.FullMethod,                 // path
				"grpc",                          // версия HTTP
				int(status.Code(err)),           // код ответа
				time.Since(start),               // время обработки запроса
				meta["user-agent"],              // user agent
			),
		)

		return result, err
	}
}
