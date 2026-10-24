package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/TarekDTM/EnergyOptimizer/internal/data"
	"github.com/TarekDTM/EnergyOptimizer/internal/validator"
	"github.com/shopspring/decimal"
)

func (app *application) createProjectsHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		UserID              int64           `json:"user_id"`
		Title               string          `json:"title"`
		MonthlyBill         decimal.Decimal `json:"monthly_bill"`
		TariffSchemeId      *int64          `json:"tariff_scheme_id,omitempty"`
		FlatElectricityRate *int64          `json:"flat_electricity_rate,omitempty"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	project := &data.Projects{
		Title:       input.Title,
		MonthlyBill: input.MonthlyBill,
		UserId:      input.UserID,
	}
	if input.TariffSchemeId != nil {
		project.TariffSchemeId = *input.TariffSchemeId
	}
	if input.FlatElectricityRate != nil {
		project.FlatElectricityRate = *input.FlatElectricityRate
	}

	v := validator.New()

	if data.ValidateProject(v, project); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Projects.Insert(project)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/projects/%d", project.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"data": project}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
func (app *application) showProjectsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	edoc, err := app.models.Edocs.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": edoc}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listProjects(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}
	v := validator.New()

	qs := r.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{
		"id", "title", "year", "-id", "-title", "-year"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	edocs, metaData, err := app.models.Edocs.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"data": edocs, "metadata": metaData}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProjects(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	project, err := app.models.Projects.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title               *string          `json:"title"`
		MonthlyBill         *decimal.Decimal `json:"monthly_bill"`
		TariffSchemeId      *int64           `json:"tariff_scheme_id,omitempty"`
		FlatElectricityRate *int64           `json:"flat_electricity_rate,omitempty"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Title != nil {

		project.Title = *input.Title
	}
	if input.MonthlyBill != nil {

		project.MonthlyBill = *input.MonthlyBill
	}
	if input.TariffSchemeId != nil {

		project.TariffSchemeId = *input.TariffSchemeId
	}
	if input.FlatElectricityRate != nil {

		project.FlatElectricityRate = *input.FlatElectricityRate
	}

	v := validator.New()

	if data.ValidateProject(v, project); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Projects.Update(project)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.errEditConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": project}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProjects(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.Edocs.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Edoc deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
