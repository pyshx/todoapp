package grpc

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	todov1 "github.com/pyshx/todoapp/gen/todo/v1"
	"github.com/pyshx/todoapp/gen/todo/v1/todov1connect"
	"github.com/pyshx/todoapp/internal/infra/postgres"
	"github.com/pyshx/todoapp/internal/usecase/taskuc"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
)

type TaskHandler struct {
	createTask       *taskuc.CreateTask
	listCompanyTasks *taskuc.ListCompanyTasks
	listMyTasks      *taskuc.ListMyTasks
	getTask          *taskuc.GetTask
	updateTask       *taskuc.UpdateTask
	deleteTask       *taskuc.DeleteTask
}

func NewTaskHandler(
	createTask *taskuc.CreateTask,
	listCompanyTasks *taskuc.ListCompanyTasks,
	listMyTasks *taskuc.ListMyTasks,
	getTask *taskuc.GetTask,
	updateTask *taskuc.UpdateTask,
	deleteTask *taskuc.DeleteTask,
) *TaskHandler {
	return &TaskHandler{
		createTask:       createTask,
		listCompanyTasks: listCompanyTasks,
		listMyTasks:      listMyTasks,
		getTask:          getTask,
		updateTask:       updateTask,
		deleteTask:       deleteTask,
	}
}

func (h *TaskHandler) CreateTask(ctx context.Context, req *connect.Request[todov1.CreateTaskRequest]) (*connect.Response[todov1.CreateTaskResponse], error) {
	actor, ok := UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	var assigneeID *id.UserID
	if req.Msg.AssigneeId != nil {
		aid, err := id.ParseUserID(*req.Msg.AssigneeId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		assigneeID = &aid
	}

	var dueDate *timestamppb.Timestamp
	if req.Msg.DueDate != nil {
		dueDate = req.Msg.DueDate
	}

	input := taskuc.CreateTaskInput{
		Title:       req.Msg.Title,
		Description: req.Msg.Description,
		AssigneeID:  assigneeID,
		Visibility:  protoToVisibility(req.Msg.Visibility),
	}
	if dueDate != nil {
		t := dueDate.AsTime()
		input.DueDate = &t
	}

	t, err := h.createTask.Execute(ctx, actor, input)
	if err != nil {
		return nil, MapError(err)
	}

	return connect.NewResponse(&todov1.CreateTaskResponse{
		Task: taskToProto(t),
	}), nil
}

func (h *TaskHandler) ListCompanyTasks(ctx context.Context, req *connect.Request[todov1.ListCompanyTasksRequest]) (*connect.Response[todov1.ListCompanyTasksResponse], error) {
	actor, ok := UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	cursor, err := postgres.DecodeCursor(req.Msg.PageToken)
	if err != nil {
		return nil, MapError(err)
	}

	input := taskuc.ListCompanyTasksInput{
		PageSize: int(req.Msg.PageSize),
		Cursor:   cursor,
	}

	result, err := h.listCompanyTasks.Execute(ctx, actor, input)
	if err != nil {
		return nil, MapError(err)
	}

	tasks := make([]*todov1.Task, len(result.Tasks))
	for i, t := range result.Tasks {
		tasks[i] = taskToProto(t)
	}

	return connect.NewResponse(&todov1.ListCompanyTasksResponse{
		Tasks:         tasks,
		NextPageToken: postgres.EncodeCursor(result.NextCursor),
	}), nil
}

func (h *TaskHandler) ListMyTasks(ctx context.Context, req *connect.Request[todov1.ListMyTasksRequest]) (*connect.Response[todov1.ListMyTasksResponse], error) {
	actor, ok := UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	cursor, err := postgres.DecodeCursor(req.Msg.PageToken)
	if err != nil {
		return nil, MapError(err)
	}

	input := taskuc.ListMyTasksInput{
		PageSize: int(req.Msg.PageSize),
		Cursor:   cursor,
	}

	result, err := h.listMyTasks.Execute(ctx, actor, input)
	if err != nil {
		return nil, MapError(err)
	}

	tasks := make([]*todov1.Task, len(result.Tasks))
	for i, t := range result.Tasks {
		tasks[i] = taskToProto(t)
	}

	return connect.NewResponse(&todov1.ListMyTasksResponse{
		Tasks:         tasks,
		NextPageToken: postgres.EncodeCursor(result.NextCursor),
	}), nil
}

func (h *TaskHandler) GetTask(ctx context.Context, req *connect.Request[todov1.GetTaskRequest]) (*connect.Response[todov1.GetTaskResponse], error) {
	actor, ok := UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	taskID, err := id.ParseTaskID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	t, err := h.getTask.Execute(ctx, actor, taskID)
	if err != nil {
		return nil, MapError(err)
	}

	return connect.NewResponse(&todov1.GetTaskResponse{
		Task: taskToProto(t),
	}), nil
}

func (h *TaskHandler) UpdateTask(ctx context.Context, req *connect.Request[todov1.UpdateTaskRequest]) (*connect.Response[todov1.UpdateTaskResponse], error) {
	actor, ok := UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	taskID, err := id.ParseTaskID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	input := taskuc.UpdateTaskInput{
		TaskID:  taskID,
		Version: int(req.Msg.Version),
	}

	if req.Msg.Title != nil {
		input.Title = req.Msg.Title
	}
	if req.Msg.Description != nil {
		input.Description = &req.Msg.Description
	}
	if req.Msg.AssigneeId != nil {
		if *req.Msg.AssigneeId == "" {
			var nilID *id.UserID
			input.AssigneeID = &nilID
		} else {
			aid, err := id.ParseUserID(*req.Msg.AssigneeId)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, err)
			}
			aidPtr := &aid
			input.AssigneeID = &aidPtr
		}
	}
	if req.Msg.DueDate != nil {
		t := req.Msg.DueDate.AsTime()
		tPtr := &t
		input.DueDate = &tPtr
	}
	if req.Msg.Visibility != nil {
		v := protoToVisibility(*req.Msg.Visibility)
		input.Visibility = &v
	}
	if req.Msg.Status != nil {
		s := protoToStatus(*req.Msg.Status)
		input.Status = &s
	}

	t, err := h.updateTask.Execute(ctx, actor, input)
	if err != nil {
		return nil, MapError(err)
	}

	return connect.NewResponse(&todov1.UpdateTaskResponse{
		Task: taskToProto(t),
	}), nil
}

func (h *TaskHandler) DeleteTask(ctx context.Context, req *connect.Request[todov1.DeleteTaskRequest]) (*connect.Response[todov1.DeleteTaskResponse], error) {
	actor, ok := UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	taskID, err := id.ParseTaskID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := h.deleteTask.Execute(ctx, actor, taskID); err != nil {
		return nil, MapError(err)
	}

	return connect.NewResponse(&todov1.DeleteTaskResponse{}), nil
}

func taskToProto(t *task.Task) *todov1.Task {
	pb := &todov1.Task{
		Id:         t.ID().String(),
		CompanyId:  t.CompanyID().String(),
		CreatorId:  t.CreatorID().String(),
		Title:      t.Title(),
		Visibility: visibilityToProto(t.Visibility()),
		Status:     statusToProto(t.Status()),
		Version:    int32(t.Version()),
		CreatedAt:  timestamppb.New(t.CreatedAt()),
		UpdatedAt:  timestamppb.New(t.UpdatedAt()),
	}

	if t.AssigneeID() != nil {
		s := t.AssigneeID().String()
		pb.AssigneeId = &s
	}
	if t.Description() != nil {
		pb.Description = t.Description()
	}
	if t.DueDate() != nil {
		pb.DueDate = timestamppb.New(*t.DueDate())
	}

	return pb
}

func visibilityToProto(v task.Visibility) todov1.Visibility {
	switch v {
	case task.VisibilityOnlyMe:
		return todov1.Visibility_VISIBILITY_ONLY_ME
	case task.VisibilityCompanyWide:
		return todov1.Visibility_VISIBILITY_COMPANY_WIDE
	default:
		return todov1.Visibility_VISIBILITY_UNSPECIFIED
	}
}

func protoToVisibility(v todov1.Visibility) task.Visibility {
	switch v {
	case todov1.Visibility_VISIBILITY_ONLY_ME:
		return task.VisibilityOnlyMe
	case todov1.Visibility_VISIBILITY_COMPANY_WIDE:
		return task.VisibilityCompanyWide
	default:
		return task.VisibilityOnlyMe
	}
}

func statusToProto(s task.Status) todov1.TaskStatus {
	switch s {
	case task.StatusTodo:
		return todov1.TaskStatus_TASK_STATUS_TODO
	case task.StatusInProgress:
		return todov1.TaskStatus_TASK_STATUS_IN_PROGRESS
	case task.StatusDone:
		return todov1.TaskStatus_TASK_STATUS_DONE
	default:
		return todov1.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func protoToStatus(s todov1.TaskStatus) task.Status {
	switch s {
	case todov1.TaskStatus_TASK_STATUS_TODO:
		return task.StatusTodo
	case todov1.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return task.StatusInProgress
	case todov1.TaskStatus_TASK_STATUS_DONE:
		return task.StatusDone
	default:
		return task.StatusTodo
	}
}

var _ todov1connect.TodoServiceHandler = (*TaskHandler)(nil)
