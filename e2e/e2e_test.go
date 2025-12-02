package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var baseURL = "http://localhost:50051"
var testUserID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

func init() {
	if url := os.Getenv("E2E_BASE_URL"); url != "" {
		baseURL = url
	}
}

func TestE2E_CreateTask(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	body := map[string]interface{}{
		"title":       "E2E Test Task",
		"description": "Created by E2E test",
		"visibility":  "VISIBILITY_COMPANY_WIDE",
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/CreateTask", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", testUserID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	task, ok := result["task"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing task field")
	}

	if task["title"] != "E2E Test Task" {
		t.Errorf("expected title 'E2E Test Task', got '%v'", task["title"])
	}
}

func TestE2E_ListCompanyTasks(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	body := map[string]interface{}{
		"page_size": 10,
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/ListCompanyTasks", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", testUserID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	tasks, ok := result["tasks"].([]interface{})
	if !ok {
		// Empty list is valid
		return
	}

	t.Logf("found %d tasks", len(tasks))
}

func TestE2E_GetTask(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	// Use a seeded task ID
	taskID := "dddddddd-dddd-dddd-dddd-dddddddddddd"

	body := map[string]interface{}{
		"id": taskID,
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/GetTask", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", testUserID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	task, ok := result["task"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing task field")
	}

	if task["id"] != taskID {
		t.Errorf("expected task id %s, got %v", taskID, task["id"])
	}
}

func TestE2E_UpdateTask(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	// First create a task
	createBody := map[string]interface{}{
		"title":       "Task to Update",
		"description": "Will be updated",
		"visibility":  "VISIBILITY_COMPANY_WIDE",
	}
	createJSON, _ := json.Marshal(createBody)

	createReq, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/CreateTask", bytes.NewReader(createJSON))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("x-user-id", testUserID)

	client := &http.Client{Timeout: 10 * time.Second}
	createResp, err := client.Do(createReq)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	defer createResp.Body.Close()

	var createResult map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	task := createResult["task"].(map[string]interface{})
	taskID := task["id"].(string)
	version := int(task["version"].(float64))

	// Now update the task
	updateBody := map[string]interface{}{
		"id":      taskID,
		"version": version,
		"title":   "Updated Task Title",
		"status":  "TASK_STATUS_IN_PROGRESS",
	}
	updateJSON, _ := json.Marshal(updateBody)

	updateReq, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/UpdateTask", bytes.NewReader(updateJSON))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("x-user-id", testUserID)

	updateResp, err := client.Do(updateReq)
	if err != nil {
		t.Fatalf("update request failed: %v", err)
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(updateResp.Body)
		t.Fatalf("expected status 200, got %d: %s", updateResp.StatusCode, body)
	}

	var updateResult map[string]interface{}
	json.NewDecoder(updateResp.Body).Decode(&updateResult)
	updatedTask := updateResult["task"].(map[string]interface{})

	if updatedTask["title"] != "Updated Task Title" {
		t.Errorf("expected updated title, got %v", updatedTask["title"])
	}
}

func TestE2E_DeleteTask(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	// First create a task
	createBody := map[string]interface{}{
		"title":      "Task to Delete",
		"visibility": "VISIBILITY_COMPANY_WIDE",
	}
	createJSON, _ := json.Marshal(createBody)

	createReq, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/CreateTask", bytes.NewReader(createJSON))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("x-user-id", testUserID)

	client := &http.Client{Timeout: 10 * time.Second}
	createResp, err := client.Do(createReq)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	defer createResp.Body.Close()

	var createResult map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	task := createResult["task"].(map[string]interface{})
	taskID := task["id"].(string)

	// Now delete the task
	deleteBody := map[string]interface{}{
		"id": taskID,
	}
	deleteJSON, _ := json.Marshal(deleteBody)

	deleteReq, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/DeleteTask", bytes.NewReader(deleteJSON))
	deleteReq.Header.Set("Content-Type", "application/json")
	deleteReq.Header.Set("x-user-id", testUserID)

	deleteResp, err := client.Do(deleteReq)
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(deleteResp.Body)
		t.Fatalf("expected status 200, got %d: %s", deleteResp.StatusCode, body)
	}

	// Verify task is deleted by trying to get it
	getBody := map[string]interface{}{
		"id": taskID,
	}
	getJSON, _ := json.Marshal(getBody)

	getReq, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/GetTask", bytes.NewReader(getJSON))
	getReq.Header.Set("Content-Type", "application/json")
	getReq.Header.Set("x-user-id", testUserID)

	getResp, err := client.Do(getReq)
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	defer getResp.Body.Close()

	// Should return not found
	if getResp.StatusCode == http.StatusOK {
		t.Error("expected task to be deleted, but it was found")
	}
}

func TestE2E_IdempotencyKey(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	idempotencyKey := "test-idempotency-key-" + time.Now().Format(time.RFC3339Nano)

	body := map[string]interface{}{
		"title":      "Idempotent Task",
		"visibility": "VISIBILITY_COMPANY_WIDE",
	}
	bodyJSON, _ := json.Marshal(body)

	// First request
	req1, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/CreateTask", bytes.NewReader(bodyJSON))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("x-user-id", testUserID)
	req1.Header.Set("Idempotency-Key", idempotencyKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("first request expected 200, got %d", resp1.StatusCode)
	}

	// Second request with same idempotency key should return conflict
	req2, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/CreateTask", bytes.NewReader(bodyJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("x-user-id", testUserID)
	req2.Header.Set("Idempotency-Key", idempotencyKey)

	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	resp2.Body.Close()

	// Should return already exists (409 conflict)
	if resp2.StatusCode == http.StatusOK {
		t.Error("expected second request to be rejected due to idempotency key")
	}
}

func TestE2E_MultiTenantIsolation(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	acmeUserID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" // Acme Corp user
	betaUserID := "cccccccc-cccc-cccc-cccc-cccccccccccc" // Beta Inc user

	// Create task as Acme user
	createBody := map[string]interface{}{
		"title":      "Acme Private Task",
		"visibility": "VISIBILITY_ONLY_ME",
	}
	createJSON, _ := json.Marshal(createBody)

	createReq, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/CreateTask", bytes.NewReader(createJSON))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("x-user-id", acmeUserID)

	client := &http.Client{Timeout: 10 * time.Second}
	createResp, err := client.Do(createReq)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	defer createResp.Body.Close()

	var createResult map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	task := createResult["task"].(map[string]interface{})
	taskID := task["id"].(string)

	// Try to access task as Beta user (different company)
	getBody := map[string]interface{}{
		"id": taskID,
	}
	getJSON, _ := json.Marshal(getBody)

	getReq, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/GetTask", bytes.NewReader(getJSON))
	getReq.Header.Set("Content-Type", "application/json")
	getReq.Header.Set("x-user-id", betaUserID)

	getResp, err := client.Do(getReq)
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	defer getResp.Body.Close()

	// Should NOT find task (different company)
	if getResp.StatusCode == http.StatusOK {
		t.Error("expected task to not be accessible by different company user")
	}
}

func TestE2E_ViewerCannotCreateTask(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	viewerUserID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" // Bob (viewer)

	body := map[string]interface{}{
		"title":      "Viewer Task",
		"visibility": "VISIBILITY_COMPANY_WIDE",
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/CreateTask", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", viewerUserID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should be forbidden (viewer cannot create)
	if resp.StatusCode == http.StatusOK {
		t.Error("expected viewer to not be able to create tasks")
	}
}

func TestE2E_Unauthenticated(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("E2E tests disabled, set E2E_ENABLED=true to run")
	}

	body := map[string]interface{}{
		"page_size": 10,
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", baseURL+"/todo.v1.TodoService/ListCompanyTasks", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	// No x-user-id header

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should be unauthenticated
	if resp.StatusCode == http.StatusOK {
		t.Error("expected unauthenticated error without user header")
	}
}
