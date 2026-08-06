package main

import (
	"context"
	"follower-service/proto"
	"follower-service/repository"
)

type followerGrpcServer struct {
	proto.UnimplementedFollowerServiceServer
	repo *repository.FollowerRepository
}

func NewFollowerGrpcServer(repo *repository.FollowerRepository) *followerGrpcServer {
	return &followerGrpcServer{repo: repo}
}

func (s *followerGrpcServer) Follow(ctx context.Context, req *proto.FollowRequest) (*proto.FollowResponse, error) {
	err := s.repo.Follow(req.FollowerId, req.FollowedId, req.FollowerUsername, req.FollowedUsername)
	if err != nil {
		return &proto.FollowResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &proto.FollowResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (s *followerGrpcServer) Unfollow(ctx context.Context, req *proto.UnfollowRequest) (*proto.UnfollowResponse, error) {
	err := s.repo.Unfollow(req.FollowerId, req.FollowedId)
	if err != nil {
		return &proto.UnfollowResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &proto.UnfollowResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (s *followerGrpcServer) GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error) {
	users, err := s.repo.GetFollowing(req.UserId)
	if err != nil {
		return &proto.GetFollowingResponse{Users: nil}, err
	}
	protoUsers := make([]*proto.UserProto, 0, len(users))
	for _, u := range users {
		protoUsers = append(protoUsers, &proto.UserProto{
			Id:       u.ID,
			Username: u.Username,
		})
	}
	return &proto.GetFollowingResponse{Users: protoUsers}, nil
}
