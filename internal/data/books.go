package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/nikitashershunov/LibraryAPI/internal/validator"
)

// Book type whose fields describe the book.
type Book struct {
	ID      int64     `json:"id"`
	Created time.Time `json:"-"`
	Title   string    `json:"title"`
	Year    int32     `json:"year,omitempty"`
	Pages   Pages     `json:"pages,omitempty"`
	Genres  []string  `json:"genres,omitempty"`
	Version int32     `json:"version"`
}

// BookModel struct wraps a sql.DB connection pool and help to work with Book struct type
// and books table in database.
type BookModel struct {
	DB *sql.DB
}

// Insert accepts a pointer to a book struct, which should contain the data for the
// new record and inserts the record into the books table.
func (b BookModel) Insert(book *Book) error {
	query := `
		INSERT INTO books (title, year, pages, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created, version`

	args := []interface{}{book.Title, book.Year, book.Pages, pq.Array(book.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return b.DB.QueryRowContext(ctx, query, args...).Scan(&book.ID, &book.Created, &book.Version)
}

// Get fetches a record from the books table and returns corresponding book struct.
// It cancels query call if SQL query does not finish during 3 seconds.
func (b BookModel) Get(id int64) (*Book, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created, title, year, pages, genres, version
		FROM books
		WHERE id = $1`

	var book Book

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Created,
		&book.Title,
		&book.Year,
		&book.Pages,
		pq.Array(&book.Genres),
		&book.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &book, nil
}

// Update updates a specific book in the books table.
func (b BookModel) Update(book *Book) error {
	query := `
		UPDATE books
		SET title = $1, year = $2, pages = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []interface{}{
		book.Title,
		book.Year,
		book.Pages,
		pq.Array(book.Genres),
		book.ID,
		book.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, args...).Scan(&book.Version)
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

// Delete is a method for deleting the record in the books table.
func (b BookModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM books
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := b.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAff == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetAll returns a list of books in the form of a string of Book type based
// on set of provided filters.
func (b BookModel) GetAll(title string, genres []string, filters Filters) ([]*Book, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created, title, year, pages, genres, version
		FROM books
		WHERE (to_tsvector('english', title) @@ plainto_tsquery('english', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}

	rows, err := b.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	books := []*Book{}

	for rows.Next() {
		var book Book

		err := rows.Scan(
			&totalRecords,
			&book.ID,
			&book.Created,
			&book.Title,
			&book.Year,
			&book.Pages,
			pq.Array(&book.Genres),
			&book.Version,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	meta := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return books, meta, nil
}

// ValidateBook run validation checks on the Book type.
func ValidateBook(v *validator.Validator, book *Book) {
	// Check book.Title
	v.Check(book.Title != "", "title", "must be provided")
	v.Check(len(book.Title) <= 500, "title", "must not be more than 500 bytes long")

	// Check book.Year
	v.Check(book.Year != 0, "year", "must be provided")
	v.Check(book.Year >= 1888, "year", "must be greater than 1888")
	v.Check(book.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	// Check book.Pages
	v.Check(book.Pages != 0, "pages", "must be provided")
	v.Check(book.Pages > 0, "pages", "must be a positive integer")

	// Check book.Genres
	v.Check(book.Genres != nil, "genres", "must be provided")
	v.Check(len(book.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(book.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(book.Genres), "genres", "must not contain duplicate values")
}
