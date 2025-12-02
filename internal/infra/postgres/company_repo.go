package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/company"
	"github.com/pyshx/todoapp/pkg/id"
)

type CompanyRepo struct {
	client *Client
}

func NewCompanyRepo(client *Client) *CompanyRepo {
	return &CompanyRepo{client: client}
}

func (r *CompanyRepo) FindByID(ctx context.Context, companyID id.CompanyID) (*company.Company, error) {
	query := `
		SELECT id, name, created_at
		FROM companies
		WHERE id = $1
	`

	row := r.client.pool.QueryRow(ctx, query, companyID.UUID())

	var dbID, name string
	var createdAt time.Time

	err := row.Scan(&dbID, &name, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NewErrNotFound("company", companyID.String())
		}
		return nil, err
	}

	parsedID, _ := id.ParseCompanyID(dbID)

	c, err := company.NewBuilder().
		ID(parsedID).
		Name(name).
		CreatedAt(createdAt).
		Build()
	if err != nil {
		return nil, err
	}

	return c, nil
}

var _ company.Repo = (*CompanyRepo)(nil)
