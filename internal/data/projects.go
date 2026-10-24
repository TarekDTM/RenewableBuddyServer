package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TarekDTM/EnergyOptimizer/internal/validator"
	"github.com/shopspring/decimal"
)

type ProjectsModel struct {
	DB *sql.DB
}

type Projects struct {
	ID                  int64           `json:"id"`
	UserId              int64           `json:"user_id"`
	Title               string          `json:"title"`
	MonthlyBill         decimal.Decimal `json:"monthly_bill"`
	TariffSchemeId      int64           `json:"tariff_scheme_Id,omitempty"`
	FlatElectricityRate int64           `json:"flat_electricity_rate,omitempty"`
	CreatedAt           time.Time       `json:"created_at,omitempty"`
	UpdatedAt           time.Time       `json:"updated_at,omitempty"`
}

func ValidateProject(v *validator.Validator, project *Projects) {

	v.Check(project.Title != "", "Title", "must be provided")
	v.Check(len(project.Title) <= 500, "Title", "must not be more than 500 bytes long")
	v.Check(project.MonthlyBill != decimal.Zero, "monthly_bill", "must be provided")
	v.Check(project.FlatElectricityRate >= 0, "flat_electricity_rate", "must not be negative")
	v.Check(project.TariffSchemeId != 0 || project.FlatElectricityRate > 0,
		"tariff_scheme_id", "either a tariff scheme ID or a positive flat electricity rate must be provided")

}

func (e ProjectsModel) Insert(project *Projects) error {
	query :=
		`INSERT INTO projects(user_id,title,monthly_bill,tariff_scheme_id,flat_electricity_rate)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING id, created_at`
	args := []interface{}{project.UserId, project.Title, project.MonthlyBill, project.TariffSchemeId, project.FlatElectricityRate}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return e.DB.QueryRowContext(ctx, query, args...).Scan(&project.ID, &project.CreatedAt)
}

func (e ProjectsModel) Get(id int64) (*Projects, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, title, monthly_bill ,tariff_scheme_id, flat_electricity_rate, created_at, 
	FROM projects
	WHERE id  = $1
	`
	var project Projects

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Title,
		&project.MonthlyBill,
		&project.TariffSchemeId,
		&project.FlatElectricityRate,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &project, nil
}
func (e ProjectsModel) GetAll(title string, genres []string, filters Filters) ([]*Projects, Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id, created_at, title, storage_url, about, version
	FROM projects
	WHERE (to_tsvector('simple',title) @@ plainto_tsquery('simple', $1)  OR $1 = '')

	ORDER BY %s %s , id ASC
	LIMIT $2 OFFSET $3 `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{title, filters.limit(), filters.offset()}
	rows, err := e.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	totalRecords := 0
	var projects = []*Projects{}
	for rows.Next() {
		var project Projects
		err := rows.Scan(
			&totalRecords,
			&project.ID,
			&project.Title,
			&project.MonthlyBill,
			&project.TariffSchemeId,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		projects = append(projects, &project)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metaData := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return projects, metaData, nil
}

func (e ProjectsModel) Update(project *Projects) error {
	query := `UPDATE projects
	SET title = $1, about = $2, storage_url = $3
	WHERE id = $4 `
	args := []interface{}{
		project.Title,
		project.TariffSchemeId,
		project.MonthlyBill,
		project.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := e.DB.QueryRowContext(ctx, query, args...).Scan(&project)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (e ProjectsModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
	DELETE FROM projects
	WHERE id =$1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := e.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
