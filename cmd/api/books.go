package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"nikitashershunov.books/internal/data"
	"nikitashershunov.books/internal/validator"
)

// getBookHandler handles the "GET /v1/books/:id" endpoint and returns a JSON response of the
// requested book record. If there is an error a JSON error is returned.
func (app *application) getBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readID(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, wrapper{"book": book}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// createBookHandler handles the "POST /v1/books" endpoint and returns a JSON response of
// the newly created book record. If there is an error a JSON error is returned.
func (app *application) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title  string     `json:"title"`
		Year   int32      `json:"year"`
		Pages  data.Pages `json:"pages"`
		Genres []string   `json:"genres"`
	}
	err := app.readJSON(w, r, &in)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	book := &data.Book{
		Title:  in.Title,
		Year:   in.Year,
		Pages:  in.Pages,
		Genres: in.Genres,
	}

	v := validator.New()
	if data.ValidateBook(v, book); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Books.Insert(book)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/books/%d", book.ID))
	err = app.writeJSON(w, http.StatusCreated, wrapper{"book": book}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateBookHandler handles "PATCH /v1/books/:id" endpoint and returns a JSON response
// of the updated book record. If there is an error a JSON error is returned.
func (app *application) updateBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readID(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(book.Version), 10) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	var in struct {
		Title  *string     `json:"title"`
		Year   *int32      `json:"year"`
		Pages  *data.Pages `json:"pages"`
		Genres []string    `json:"genres"`
	}

	err = app.readJSON(w, r, &in)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if in.Title != nil {
		book.Title = *in.Title
	}
	if in.Year != nil {
		book.Year = *in.Year
	}
	if in.Pages != nil {
		book.Pages = *in.Pages
	}
	if in.Genres != nil {
		book.Genres = in.Genres
	}

	v := validator.New()
	if data.ValidateBook(v, book); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Books.Update(book)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, wrapper{"book": book}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteBookHandler handles "DELETE /v1/books/:id" endpoint and returns a 200 OK status code
// with a success message in a JSON response. If there is an error a JSON formatted error is returned.
func (app *application) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readID(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Books.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, wrapper{"message": "book successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listBooksHandler handles the "GET /v1/books" endpoint and returns a JSON response of
// the array of book records based on the query string parameters (provided filters).
// If there is an error a JSON error is returned.
func (app *application) listBooksHandler(w http.ResponseWriter, r *http.Request) {
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

	input.Filters.SortSafelist = []string{
		// ascending sort values
		"id", "title", "year", "pages",
		// descending sort values
		"-id", "-title", "-year", "-pages",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	books, meta, err := app.models.Books.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, wrapper{"books": books, "metadata": meta}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
