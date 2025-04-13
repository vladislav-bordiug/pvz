package grpch

import (
	"context"

	pb "pvz/internal/pb/pvz_v1"
	"pvz/internal/services"
)

type GrpcServer struct {
	pb.UnimplementedPVZServiceServer
	services services.ServiceInterface
}

func NewGrpcServer(services services.ServiceInterface) *GrpcServer {
	return &GrpcServer{services: services}
}

func (s *GrpcServer) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	pvzs, err := s.services.GetPVZ(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetPVZListResponse{
		Pvzs: pvzs,
	}, nil
}
