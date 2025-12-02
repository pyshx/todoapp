package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
)

type TaskRepo struct {
	client *Client
}

func NewTaskRepo(client *Client) *TaskRepo {
	return &TaskRepo{client: client}
}

func (r *TaskRepo) Create(ctx context.Context, t *task.Task) error {
	query := `
		INSERT INTO tasks (id, company_id, creator_id, assignee_id, title, description, due_date, visibility, status, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	var assigneeID interface{}
	if t.AssigneeID() != nil {
		assigneeID = t.AssigneeID().UUID()
	}

	_, err := r.client.pool.Exec(ctx, query,
		t.ID().UUID(),
		t.CompanyID().UUID(),
		t.CreatorID().UUID(),
		assigneeID,
		t.Title(),
		t.Description(),
		t.DueDate(),
		t.Visibility().String(),
		t.Status().String(),
		t.Version(),
		t.CreatedAt(),
		t.UpdatedAt(),
	)
	return err
}

func (r *TaskRepo) FindByID(ctx context.Context, taskID id.TaskID) (*task.Task, error) {
	query := `
		SELECT id, company_id, creator_id, assignee_id, title, description, due_date, visibility, status, version, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	return r.scanTask(ctx, r.client.pool.QueryRow(ctx, query, taskID.UUID()), taskID.String())
}

func (r *TaskRepo) FindByIDForCompany(ctx context.Context, taskID id.TaskID, companyID id.CompanyID) (*task.Task, error) {
	query := `
		SELECT id, company_id, creator_id, assignee_id, title, description, due_date, visibility, status, version, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND company_id = $2
	`
	return r.scanTask(ctx, r.client.pool.QueryRow(ctx, query, taskID.UUID(), companyID.UUID()), taskID.String())
}

func (r *TaskRepo) ListByCompany(ctx context.Context, companyID id.CompanyID, opts task.ListOptions) (*task.ListResult, error) {
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var rows pgx.Rows
	var err error

	if opts.Cursor != nil {
		query := `
			SELECT id, company_id, creator_id, assignee_id, title, description, due_date, visibility, status, version, created_at, updated_at
			FROM tasks
			WHERE company_id = $1 AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		`
		rows, err = r.client.pool.Query(ctx, query, companyID.UUID(), opts.Cursor.CreatedAt, opts.Cursor.ID.UUID(), pageSize+1)
	} else {
		query := `
			SELECT id, company_id, creator_id, assignee_id, title, description, due_date, visibility, status, version, created_at, updated_at
			FROM tasks
			WHERE company_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`
		rows, err = r.client.pool.Query(ctx, query, companyID.UUID(), pageSize+1)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTaskList(rows, pageSize)
}

func (r *TaskRepo) ListByAssignee(ctx context.Context, companyID id.CompanyID, assigneeID id.UserID, opts task.ListOptions) (*task.ListResult, error) {
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var rows pgx.Rows
	var err error

	if opts.Cursor != nil {
		query := `
			SELECT id, company_id, creator_id, assignee_id, title, description, due_date, visibility, status, version, created_at, updated_at
			FROM tasks
			WHERE company_id = $1 AND assignee_id = $2 AND (created_at, id) < ($3, $4)
			ORDER BY created_at DESC, id DESC
			LIMIT $5
		`
		rows, err = r.client.pool.Query(ctx, query, companyID.UUID(), assigneeID.UUID(), opts.Cursor.CreatedAt, opts.Cursor.ID.UUID(), pageSize+1)
	} else {
		query := `
			SELECT id, company_id, creator_id, assignee_id, title, description, due_date, visibility, status, version, created_at, updated_at
			FROM tasks
			WHERE company_id = $1 AND assignee_id = $2
			ORDER BY created_at DESC, id DESC
			LIMIT $3
		`
		rows, err = r.client.pool.Query(ctx, query, companyID.UUID(), assigneeID.UUID(), pageSize+1)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTaskList(rows, pageSize)
}

func (r *TaskRepo) Update(ctx context.Context, t *task.Task, expectedVersion int) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, assignee_id = $3, due_date = $4, visibility = $5, status = $6, version = $7, updated_at = $8
		WHERE id = $9 AND company_id = $10 AND version = $11
	`

	var assigneeID interface{}
	if t.AssigneeID() != nil {
		assigneeID = t.AssigneeID().UUID()
	}

	result, err := r.client.pool.Exec(ctx, query,
		t.Title(),
		t.Description(),
		assigneeID,
		t.DueDate(),
		t.Visibility().String(),
		t.Status().String(),
		t.Version(),
		t.UpdatedAt(),
		t.ID().UUID(),
		t.CompanyID().UUID(),
		expectedVersion,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		existing, err := r.FindByIDForCompany(ctx, t.ID(), t.CompanyID())
		if err != nil {
			if apperr.IsNotFound(err) {
				return err
			}
			return err
		}
		return apperr.NewErrVersionMismatch(expectedVersion, existing.Version())
	}

	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, taskID id.TaskID, companyID id.CompanyID) error {
	query := `DELETE FROM tasks WHERE id = $1 AND company_id = $2`

	result, err := r.client.pool.Exec(ctx, query, taskID.UUID(), companyID.UUID())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return apperr.NewErrNotFound("task", taskID.String())
	}

	return nil
}

func (r *TaskRepo) scanTask(ctx context.Context, row pgx.Row, taskIDStr string) (*task.Task, error) {
	var dbID, dbCompanyID, dbCreatorID string
	var dbAssigneeID *string
	var title string
	var description *string
	var dueDate *time.Time
	var visibility, status string
	var version int
	var createdAt, updatedAt time.Time

	err := row.Scan(&dbID, &dbCompanyID, &dbCreatorID, &dbAssigneeID, &title, &description, &dueDate, &visibility, &status, &version, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NewErrNotFound("task", taskIDStr)
		}
		return nil, err
	}

	return r.buildTask(dbID, dbCompanyID, dbCreatorID, dbAssigneeID, title, description, dueDate, visibility, status, version, createdAt, updatedAt)
}

func (r *TaskRepo) scanTaskList(rows pgx.Rows, pageSize int) (*task.ListResult, error) {
	var tasks []*task.Task

	for rows.Next() {
		var dbID, dbCompanyID, dbCreatorID string
		var dbAssigneeID *string
		var title string
		var description *string
		var dueDate *time.Time
		var visibility, status string
		var version int
		var createdAt, updatedAt time.Time

		err := rows.Scan(&dbID, &dbCompanyID, &dbCreatorID, &dbAssigneeID, &title, &description, &dueDate, &visibility, &status, &version, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		t, err := r.buildTask(dbID, dbCompanyID, dbCreatorID, dbAssigneeID, title, description, dueDate, visibility, status, version, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &task.ListResult{}

	if len(tasks) > pageSize {
		tasks = tasks[:pageSize]
		lastTask := tasks[len(tasks)-1]
		result.NextCursor = &task.PageCursor{
			CreatedAt: lastTask.CreatedAt(),
			ID:        lastTask.ID(),
		}
	}

	result.Tasks = tasks
	return result, nil
}

func (r *TaskRepo) buildTask(dbID, dbCompanyID, dbCreatorID string, dbAssigneeID *string, title string, description *string, dueDate *time.Time, visibility, status string, version int, createdAt, updatedAt time.Time) (*task.Task, error) {
	parsedID, _ := id.ParseTaskID(dbID)
	parsedCompanyID, _ := id.ParseCompanyID(dbCompanyID)
	parsedCreatorID, _ := id.ParseUserID(dbCreatorID)

	var parsedAssigneeID *id.UserID
	if dbAssigneeID != nil {
		aid, _ := id.ParseUserID(*dbAssigneeID)
		parsedAssigneeID = &aid
	}

	parsedVisibility, _ := task.ParseVisibility(visibility)
	parsedStatus, _ := task.ParseStatus(status)

	return task.NewBuilder().
		ID(parsedID).
		CompanyID(parsedCompanyID).
		CreatorID(parsedCreatorID).
		AssigneeID(parsedAssigneeID).
		Title(title).
		Description(description).
		DueDate(dueDate).
		Visibility(parsedVisibility).
		Status(parsedStatus).
		Version(version).
		CreatedAt(createdAt).
		UpdatedAt(updatedAt).
		Build()
}

func EncodeCursor(cursor *task.PageCursor) string {
	if cursor == nil {
		return ""
	}
	data, _ := json.Marshal(map[string]interface{}{
		"created_at": cursor.CreatedAt,
		"id":         cursor.ID.String(),
	})
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeCursor(token string) (*task.PageCursor, error) {
	if token == "" {
		return nil, nil
	}

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, apperr.NewErrInvalidInput("page_token", "invalid cursor format")
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, apperr.NewErrInvalidInput("page_token", "invalid cursor format")
	}

	createdAtStr, ok := m["created_at"].(string)
	if !ok {
		return nil, apperr.NewErrInvalidInput("page_token", "invalid cursor format")
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return nil, apperr.NewErrInvalidInput("page_token", "invalid cursor format")
	}

	idStr, ok := m["id"].(string)
	if !ok {
		return nil, apperr.NewErrInvalidInput("page_token", "invalid cursor format")
	}
	taskID, err := id.ParseTaskID(idStr)
	if err != nil {
		return nil, apperr.NewErrInvalidInput("page_token", "invalid cursor format")
	}

	return &task.PageCursor{
		CreatedAt: createdAt,
		ID:        taskID,
	}, nil
}

var _ task.Repo = (*TaskRepo)(nil)
