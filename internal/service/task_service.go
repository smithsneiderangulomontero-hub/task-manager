package service

import (
	"context"
	"fmt"
	"html"
	"strings"

	"task-manager/internal/domain"
	"task-manager/internal/repository"
)

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) GetTask(ctx context.Context, id int64) (*domain.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) ListTasks(ctx context.Context) ([]*domain.Task, error) {
	return s.repo.List(ctx)
}

func (s *TaskService) CreateTask(ctx context.Context, title, description string) (*domain.Task, error) {
	title = strings.TrimSpace(title)
	title = html.EscapeString(title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrBadRequest)
	}
	if len(title) > 255 {
		return nil, fmt.Errorf("%w: title must not exceed 255 characters", domain.ErrBadRequest)
	}
	task := &domain.Task{
		Title:       title,
		Description: strings.TrimSpace(description),
		Status:      domain.StatusTodo,
	}
	return s.repo.Create(ctx, task)
}

func (s *TaskService) UpdateTask(ctx context.Context, id int64, title, description string, status domain.TaskStatus) (*domain.Task, error) {
	title = strings.TrimSpace(title)
	title = html.EscapeString(title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrBadRequest)
	}
	if len(title) > 255 {
		return nil, fmt.Errorf("%w: title must not exceed 255 characters", domain.ErrBadRequest)
	}
	valid := map[domain.TaskStatus]bool{
		domain.StatusTodo: true, domain.StatusInProgress: true, domain.StatusDone: true,
	}
	if !valid[status] {
		return nil, fmt.Errorf("%w: invalid status %q", domain.ErrBadRequest, status)
	}
	task := &domain.Task{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(description),
		Status:      status,
	}
	return s.repo.Update(ctx, task)
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
