package grpc

import (
	"context"
	"go_micro_lab/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskServer struct {
	UnimplementedTaskServiceServer
	usecase *usecase.TaskUsecase
}

func NewTaskServer(usecase *usecase.TaskUsecase) *TaskServer {
	return &TaskServer{usecase: usecase}
}

func (s *TaskServer) CreateTask(ctx context.Context, req *CreateTaskRequest) (*CreateTaskResponse, error) {
	task, err := s.usecase.CreateTask(ctx, req.Payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &CreateTaskResponse{
		Id:     task.ID,
		Status: task.Status,
	}, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *GetTaskRequest) (*GetTaskResponse, error) {
	task, err := s.usecase.GetTask(ctx, req.Id)
	if err != nil {
		if err.Error() == "task not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &GetTaskResponse{
		Id:      task.ID,
		Payload: task.Payload,
		Status:  task.Status,
	}, nil
}