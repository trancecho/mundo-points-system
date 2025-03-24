package meta

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetMetadata(ctx context.Context) (metadata.MD, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(400, "failed to extract metadata from context")
	}
	return md, nil
}
