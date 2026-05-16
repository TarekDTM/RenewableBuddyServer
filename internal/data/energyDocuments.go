package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TarekDTM/EnergyOptimizer/internal/validator"
)

type EdocModel struct {
	DB *sql.DB
}

type EnergyDocuments struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	StorageURL  string    `json:"storage_url"`
	Version     int64     `json:"version"`
}

func ValidateEDoc(v *validator.Validator, doc *EnergyDocuments) {

	v.Check(doc.Title != "", "Title", "must be provided")
	v.Check(len(doc.Title) <= 500, "Title", "must not be more than 500 bytes long")
	v.Check(doc.Description != "", "description", "must be provided")
	v.Check(doc.StorageURL != "", "storage_url", "must be provided")

}

func (e EdocModel) Insert(energyDocument *EnergyDocuments) error {
	query :=
		`INSERT INTO energy_document(title,about,storage_url)
	VALUES ($1,$2,$3)
	RETURNING id, created_at,version`
	args := []interface{}{energyDocument.Title, energyDocument.Description, energyDocument.StorageURL}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return e.DB.QueryRowContext(ctx, query, args...).Scan(&energyDocument.ID, &energyDocument.CreatedAt, &energyDocument.Version)
}

func (e EdocModel) Get(id int64) (*EnergyDocuments, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, title, about ,storage_url, created_at, version
	FROM energy_document
	WHERE id  = $1
	`
	var eDoc EnergyDocuments

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&eDoc.ID,
		&eDoc.Title,
		&eDoc.Description,
		&eDoc.StorageURL,
		&eDoc.CreatedAt,
		&eDoc.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &eDoc, nil
}
func (e EdocModel) GetAll(title string, genres []string, filters Filters) ([]*EnergyDocuments, Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id, created_at, title, storage_url, about, version
	FROM energy_document
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
	var edocs = []*EnergyDocuments{}
	for rows.Next() {
		var doc EnergyDocuments
		err := rows.Scan(
			&totalRecords,
			&doc.ID,
			&doc.CreatedAt,
			&doc.Title,
			&doc.StorageURL,
			&doc.Description,
			&doc.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		edocs = append(edocs, &doc)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metaData := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return edocs, metaData, nil
}

func (e EdocModel) Update(energyDocument *EnergyDocuments) error {
	query := `UPDATE energy_document
	SET title = $1, about = $2, storage_url = $3, version = version + 1
	WHERE id = $4 AND version = $5
	RETURNING version`
	args := []interface{}{
		energyDocument.Title,
		energyDocument.Description,
		energyDocument.StorageURL,
		energyDocument.ID,
		energyDocument.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := e.DB.QueryRowContext(ctx, query, args...).Scan(&energyDocument.Version)
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

func (e EdocModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
	DELETE FROM energy_document
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
