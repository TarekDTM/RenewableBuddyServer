package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/TarekDTM/EnergyOptimizer/internal/data"
	"github.com/TarekDTM/EnergyOptimizer/internal/validator"
)

func (app *application) createEnergyDocumentHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		StorageURL  string `json:"storage_url"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	eDoc := &data.EnergyDocuments{
		Title:       input.Name,
		Description: input.Description,
		StorageURL:  input.StorageURL,
	}

	v := validator.New()

	if data.ValidateEDoc(v, eDoc); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Edocs.Insert(eDoc)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", eDoc.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"edoc": eDoc}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
func (app *application) showEnergyDocumentHandler(w http.ResponseWriter, r *http.Request) {
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

	err = app.writeJSON(w, http.StatusOK, envelope{"doc": edoc}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listEnergyDocuments(w http.ResponseWriter, r *http.Request) {
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
	err = app.writeJSON(w, http.StatusOK, envelope{"edocs": edocs, "metadata": metaData}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateEnergyDocument(w http.ResponseWriter, r *http.Request) {
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

	var input struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		StorageURL  *string `json:"storage_url"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Name != nil {

		edoc.Title = *input.Name
	}
	if input.Description != nil {

		edoc.Description = *input.Description
	}
	if input.StorageURL != nil {

		edoc.StorageURL = *input.StorageURL
	}

	v := validator.New()

	if data.ValidateEDoc(v, edoc); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Edocs.Update(edoc)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.errEditConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"edoc": edoc}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteEnergyDocument(w http.ResponseWriter, r *http.Request) {
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
