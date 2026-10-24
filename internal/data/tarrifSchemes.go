package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TarekDTM/EnergyOptimizer/internal/validator"
)

type TarrifSchemesModel struct {
	DB *sql.DB
}

type TarrifSchemes struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Currency      string    `json:"currency"`
	EffectiveFrom time.Time `json:"effective_from,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

func ValidateTarrifSchemes(v *validator.Validator, tarrif_scheme *TarrifSchemes) {

	v.Check(tarrif_scheme.Title != "", "Title", "must be provided")
	v.Check(len(tarrif_scheme.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(tarrif_scheme.Currency != "", "currency", "must be provided")

}

func (e TarrifSchemesModel) Insert(tarrif_scheme *TarrifSchemes) error {
	query :=
		`INSERT INTO tarrif_schemes(title,currency,effective_from)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING id, created_at`
	args := []interface{}{tarrif_scheme.Title, tarrif_scheme.Currency, tarrif_scheme.EffectiveFrom}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return e.DB.QueryRowContext(ctx, query, args...).Scan(&tarrif_scheme.ID, &tarrif_scheme.CreatedAt)
}

func (e TarrifSchemesModel) Get(id int64) (*TarrifSchemes, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, title, about ,storage_url, created_at, version
	FROM projects
	WHERE id  = $1
	`
	var tarrif_scheme TarrifSchemes

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&tarrif_scheme.ID,
		&tarrif_scheme.Title,
		&tarrif_scheme.Currency,
		&tarrif_scheme.EffectiveFrom,
		&tarrif_scheme.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &tarrif_scheme, nil
}
func (e TarrifSchemesModel) GetAll(title string, genres []string, filters Filters) ([]*Projects, Metadata, error) {
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

func (e TarrifSchemesModel) Update(tarrif_scheme *TarrifSchemes) error {
	query := `UPDATE projects
	SET title = $1, about = $2, storage_url = $3
	WHERE id = $4 `
	args := []interface{}{
		tarrif_scheme.Title,
		tarrif_scheme.Currency,
		tarrif_scheme.EffectiveFrom,
		tarrif_scheme.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := e.DB.QueryRowContext(ctx, query, args...).Scan(&tarrif_scheme)
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

func Delete(e TarrifSchemesModel, id int64) error {
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
