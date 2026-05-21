package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"task-manager/internal/domain"
	"task-manager/internal/service"
)

type MockTaskRepository struct{ mock.Mock }

func (m *MockTaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}
func (m *MockTaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}
func (m *MockTaskRepository) Create(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}
func (m *MockTaskRepository) Update(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}
func (m *MockTaskRepository) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func TestTaskService_CreateTask(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		wantErr     bool
		errContains string
	}{
		{name: "valid task", title: "Fix the bug", wantErr: false},
		{name: "empty title rejected", title: "   ", wantErr: true, errContains: "title is required"},
		{name: "title too long rejected", title: string(make([]byte, 256)), wantErr: true, errContains: "255 characters"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockTaskRepository)
			if !tc.wantErr {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).
					Return(&domain.Task{ID: 1, Title: tc.title, Status: domain.StatusTodo}, nil)
			}
			svc := service.NewTaskService(repo)
			got, err := svc.CreateTask(context.Background(), tc.title, "")
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.StatusTodo, got.Status)
			repo.AssertExpectations(t)
		})
	}
}

func TestTaskService_GetTask(t *testing.T) {
	t.Run("returns task when found", func(t *testing.T) {
		repo := new(MockTaskRepository)
		expected := &domain.Task{ID: 1, Title: "Test", Status: domain.StatusTodo}
		repo.On("GetByID", mock.Anything, int64(1)).Return(expected, nil)
		got, err := service.NewTaskService(repo).GetTask(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
		repo.AssertExpectations(t)
	})

	t.Run("returns ErrNotFound", func(t *testing.T) {
		repo := new(MockTaskRepository)
		repo.On("GetByID", mock.Anything, int64(99)).Return(nil, domain.ErrNotFound)
		got, err := service.NewTaskService(repo).GetTask(context.Background(), 99)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	t.Run("deletes successfully", func(t *testing.T) {
		repo := new(MockTaskRepository)
		repo.On("Delete", mock.Anything, int64(1)).Return(nil)
		err := service.NewTaskService(repo).DeleteTask(context.Background(), 1)
		require.NoError(t, err)
	})

	t.Run("returns ErrNotFound", func(t *testing.T) {
		repo := new(MockTaskRepository)
		repo.On("Delete", mock.Anything, int64(99)).Return(domain.ErrNotFound)
		err := service.NewTaskService(repo).DeleteTask(context.Background(), 99)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestTaskService_ListTasks(t *testing.T) {
	t.Run("returns list of tasks", func(t *testing.T) {
		repo := new(MockTaskRepository)
		expected := []*domain.Task{
			{ID: 1, Title: "Task 1", Status: domain.StatusTodo},
			{ID: 2, Title: "Task 2", Status: domain.StatusDone},
		}
		repo.On("List", mock.Anything).Return(expected, nil)

		got, err := service.NewTaskService(repo).ListTasks(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)
		repo.AssertExpectations(t)
	})

	t.Run("returns empty list when no tasks", func(t *testing.T) {
		repo := new(MockTaskRepository)
		repo.On("List", mock.Anything).Return([]*domain.Task{}, nil)

		got, err := service.NewTaskService(repo).ListTasks(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		status      domain.TaskStatus
		wantErr     bool
		errContains string
	}{
		{
			name:   "valid update",
			title:  "Updated title",
			status: domain.StatusInProgress,
		},
		{
			name:        "empty title rejected",
			title:       "  ",
			status:      domain.StatusTodo,
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name:        "invalid status rejected",
			title:       "Valid title",
			status:      "unknown",
			wantErr:     true,
			errContains: "invalid status",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockTaskRepository)
			if !tc.wantErr {
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).
					Return(&domain.Task{ID: 1, Title: tc.title, Status: tc.status}, nil)
			}

			svc := service.NewTaskService(repo)
			got, err := svc.UpdateTask(context.Background(), 1, tc.title, "", tc.status)

			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.status, got.Status)
			repo.AssertExpectations(t)
		})
	}
}
