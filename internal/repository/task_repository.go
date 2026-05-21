package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager/internal/domain"
)

type TaskRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Task, error)
	List(ctx context.Context) ([]*domain.Task, error)
	Create(ctx context.Context, task *domain.Task) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) (*domain.Task, error)
	Delete(ctx context.Context, id int64) error
}

type postgresTaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) TaskRepository {
	return &postgresTaskRepository{db: db}
}

func (r *postgresTaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	query := `SELECT id, title, description, status, created_at, updated_at
	          FROM tasks WHERE id = $1`
	var t domain.Task
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get task by id: %w", err)
	}
	return &t, nil
}

func (r *postgresTaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	query := `SELECT id, title, description, status, created_at, updated_at
	          FROM tasks ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task row: %w", err)
		}
		tasks = append(tasks, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate task rows: %w", err)
	}
	return tasks, nil
}

func (r *postgresTaskRepository) Create(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	query := `INSERT INTO tasks (title, description, status)
	          VALUES ($1, $2, $3)
	          RETURNING id, title, description, status, created_at, updated_at`
	var t domain.Task
	err := r.db.QueryRow(ctx, query, task.Title, task.Description, task.Status).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return &t, nil
}

func (r *postgresTaskRepository) Update(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	query := `UPDATE tasks SET title=$1, description=$2, status=$3, updated_at=NOW()
	          WHERE id=$4
	          RETURNING id, title, description, status, created_at, updated_at`
	var t domain.Task
	err := r.db.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.ID).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update task: %w", err)
	}
	return &t, nil
}

func (r *postgresTaskRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
