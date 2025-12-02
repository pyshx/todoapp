package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/user"
)

type UserRepo struct {
	client *Client
}

func NewUserRepo(client *Client) *UserRepo {
	return &UserRepo{client: client}
}

func (r *UserRepo) FindByID(ctx context.Context, userID id.UserID) (*user.User, error) {
	query := `
		SELECT id, company_id, email, role, created_at
		FROM users
		WHERE id = $1
	`

	row := r.client.pool.QueryRow(ctx, query, userID.UUID())

	var dbID, dbCompanyID string
	var email, role string
	var createdAt interface{}

	err := row.Scan(&dbID, &dbCompanyID, &email, &role, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NewErrNotFound("user", userID.String())
		}
		return nil, err
	}

	parsedRole, ok := user.ParseRole(role)
	if !ok {
		parsedRole = user.RoleViewer
	}

	companyID, _ := id.ParseCompanyID(dbCompanyID)
	parsedID, _ := id.ParseUserID(dbID)

	u, err := user.NewBuilder().
		ID(parsedID).
		CompanyID(companyID).
		Email(email).
		Role(parsedRole).
		Build()
	if err != nil {
		return nil, err
	}

	return u, nil
}

var _ user.Repo = (*UserRepo)(nil)
