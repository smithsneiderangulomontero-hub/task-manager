package handler

import "task-manager/internal/service"

type Handler struct {
	taskService *service.TaskService
}

func New(taskService *service.TaskService) *Handler {
	return &Handler{taskService: taskService}
}
