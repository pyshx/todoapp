package di

import (
	"context"
	"log/slog"

	grpcserver "github.com/pyshx/todoapp/internal/infra/grpc"
	"github.com/pyshx/todoapp/internal/infra/postgres"
	"github.com/pyshx/todoapp/internal/usecase/taskuc"
	"github.com/pyshx/todoapp/pkg/user"
)

type Container struct {
	DBClient    *postgres.Client
	UserRepo    user.Repo
	TaskHandler *grpcserver.TaskHandler
	Server      *grpcserver.Server
}

func New(ctx context.Context, databaseURL string, grpcPort int, logger *slog.Logger) (*Container, error) {
	dbClient, err := postgres.NewClient(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	userRepo := postgres.NewUserRepo(dbClient)
	taskRepo := postgres.NewTaskRepo(dbClient)

	createTask := taskuc.NewCreateTask(taskRepo, userRepo)
	listCompanyTasks := taskuc.NewListCompanyTasks(taskRepo)
	listMyTasks := taskuc.NewListMyTasks(taskRepo)
	getTask := taskuc.NewGetTask(taskRepo)
	updateTask := taskuc.NewUpdateTask(taskRepo, userRepo)
	deleteTask := taskuc.NewDeleteTask(taskRepo)

	taskHandler := grpcserver.NewTaskHandler(
		createTask,
		listCompanyTasks,
		listMyTasks,
		getTask,
		updateTask,
		deleteTask,
	)

	server := grpcserver.NewServer(grpcPort, taskHandler, userRepo, logger)

	return &Container{
		DBClient:    dbClient,
		UserRepo:    userRepo,
		TaskHandler: taskHandler,
		Server:      server,
	}, nil
}

func (c *Container) Close() {
	if c.DBClient != nil {
		c.DBClient.Close()
	}
}
