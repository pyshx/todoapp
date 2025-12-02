package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pyshx/todoapp/internal/infra/postgres"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
)

func setupTestDB(t *testing.T) *postgres.Client {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	client, err := postgres.NewClient(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	return client
}

func TestTaskRepo_CreateAndFindByID(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	repo := postgres.NewTaskRepo(client)

	companyID, _ := id.ParseCompanyID("11111111-1111-1111-1111-111111111111")
	creatorID, _ := id.ParseUserID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	taskID := id.NewTaskID()
	now := time.Now().Truncate(time.Microsecond)

	newTask, _ := task.NewBuilder().
		ID(taskID).
		CompanyID(companyID).
		CreatorID(creatorID).
		Title("Integration Test Task").
		Visibility(task.VisibilityCompanyWide).
		Status(task.StatusTodo).
		Version(1).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	if err := repo.Create(ctx, newTask); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	found, err := repo.FindByID(ctx, taskID)
	if err != nil {
		t.Fatalf("failed to find task: %v", err)
	}

	if found.Title() != "Integration Test Task" {
		t.Errorf("expected title 'Integration Test Task', got '%s'", found.Title())
	}
	if found.Status() != task.StatusTodo {
		t.Errorf("expected status todo, got %s", found.Status())
	}

	// Cleanup
	repo.Delete(ctx, taskID, companyID)
}

func TestTaskRepo_FindByIDForCompany(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	repo := postgres.NewTaskRepo(client)

	companyID, _ := id.ParseCompanyID("11111111-1111-1111-1111-111111111111")
	otherCompanyID, _ := id.ParseCompanyID("22222222-2222-2222-2222-222222222222")
	creatorID, _ := id.ParseUserID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	taskID := id.NewTaskID()
	now := time.Now().Truncate(time.Microsecond)

	newTask, _ := task.NewBuilder().
		ID(taskID).
		CompanyID(companyID).
		CreatorID(creatorID).
		Title("Company Isolation Test").
		Visibility(task.VisibilityCompanyWide).
		Status(task.StatusTodo).
		Version(1).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	if err := repo.Create(ctx, newTask); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Should find with correct company
	found, err := repo.FindByIDForCompany(ctx, taskID, companyID)
	if err != nil {
		t.Fatalf("failed to find task with correct company: %v", err)
	}
	if found == nil {
		t.Error("expected to find task")
	}

	// Should NOT find with wrong company
	_, err = repo.FindByIDForCompany(ctx, taskID, otherCompanyID)
	if err == nil {
		t.Error("expected error when finding task with wrong company")
	}

	// Cleanup
	repo.Delete(ctx, taskID, companyID)
}

func TestTaskRepo_Update(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	repo := postgres.NewTaskRepo(client)

	companyID, _ := id.ParseCompanyID("11111111-1111-1111-1111-111111111111")
	creatorID, _ := id.ParseUserID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	taskID := id.NewTaskID()
	now := time.Now().Truncate(time.Microsecond)

	originalTask, _ := task.NewBuilder().
		ID(taskID).
		CompanyID(companyID).
		CreatorID(creatorID).
		Title("Original Title").
		Visibility(task.VisibilityCompanyWide).
		Status(task.StatusTodo).
		Version(1).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	if err := repo.Create(ctx, originalTask); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Update the task
	newTitle := "Updated Title"
	update := task.Update{
		Title: &newTitle,
	}
	updatedTask := originalTask.ApplyUpdate(update, time.Now())

	if err := repo.Update(ctx, updatedTask, 1); err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	// Verify update
	found, err := repo.FindByID(ctx, taskID)
	if err != nil {
		t.Fatalf("failed to find updated task: %v", err)
	}

	if found.Title() != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", found.Title())
	}
	if found.Version() != 2 {
		t.Errorf("expected version 2, got %d", found.Version())
	}

	// Cleanup
	repo.Delete(ctx, taskID, companyID)
}

func TestTaskRepo_OptimisticLocking(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	repo := postgres.NewTaskRepo(client)

	companyID, _ := id.ParseCompanyID("11111111-1111-1111-1111-111111111111")
	creatorID, _ := id.ParseUserID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	taskID := id.NewTaskID()
	now := time.Now().Truncate(time.Microsecond)

	originalTask, _ := task.NewBuilder().
		ID(taskID).
		CompanyID(companyID).
		CreatorID(creatorID).
		Title("Optimistic Lock Test").
		Visibility(task.VisibilityCompanyWide).
		Status(task.StatusTodo).
		Version(1).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	if err := repo.Create(ctx, originalTask); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// First update should succeed
	newTitle := "First Update"
	update := task.Update{Title: &newTitle}
	updatedTask := originalTask.ApplyUpdate(update, time.Now())

	if err := repo.Update(ctx, updatedTask, 1); err != nil {
		t.Fatalf("first update should succeed: %v", err)
	}

	// Second update with stale version should fail
	staleTitle := "Stale Update"
	staleUpdate := task.Update{Title: &staleTitle}
	staleTask := originalTask.ApplyUpdate(staleUpdate, time.Now())

	err := repo.Update(ctx, staleTask, 1) // Using stale version 1
	if err == nil {
		t.Error("expected version mismatch error for stale update")
	}

	// Cleanup
	repo.Delete(ctx, taskID, companyID)
}

func TestTaskRepo_ListByCompany(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	repo := postgres.NewTaskRepo(client)

	companyID, _ := id.ParseCompanyID("11111111-1111-1111-1111-111111111111")
	creatorID, _ := id.ParseUserID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	// Create test tasks
	var taskIDs []id.TaskID
	for i := 0; i < 3; i++ {
		taskID := id.NewTaskID()
		taskIDs = append(taskIDs, taskID)
		now := time.Now().Truncate(time.Microsecond)

		newTask, _ := task.NewBuilder().
			ID(taskID).
			CompanyID(companyID).
			CreatorID(creatorID).
			Title("List Test Task").
			Visibility(task.VisibilityCompanyWide).
			Status(task.StatusTodo).
			Version(1).
			CreatedAt(now).
			UpdatedAt(now).
			Build()

		if err := repo.Create(ctx, newTask); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	// List tasks
	result, err := repo.ListByCompany(ctx, companyID, task.ListOptions{PageSize: 10})
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(result.Tasks) < 3 {
		t.Errorf("expected at least 3 tasks, got %d", len(result.Tasks))
	}

	// Cleanup
	for _, taskID := range taskIDs {
		repo.Delete(ctx, taskID, companyID)
	}
}

func TestTaskRepo_Delete(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	repo := postgres.NewTaskRepo(client)

	companyID, _ := id.ParseCompanyID("11111111-1111-1111-1111-111111111111")
	creatorID, _ := id.ParseUserID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	taskID := id.NewTaskID()
	now := time.Now().Truncate(time.Microsecond)

	newTask, _ := task.NewBuilder().
		ID(taskID).
		CompanyID(companyID).
		CreatorID(creatorID).
		Title("Delete Test Task").
		Visibility(task.VisibilityCompanyWide).
		Status(task.StatusTodo).
		Version(1).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	if err := repo.Create(ctx, newTask); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Delete the task
	if err := repo.Delete(ctx, taskID, companyID); err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	// Verify deletion
	_, err := repo.FindByID(ctx, taskID)
	if err == nil {
		t.Error("expected error when finding deleted task")
	}
}
