package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"nikitashershunov.books/internal/validator"

	"github.com/julienschmidt/httprouter"
)

// define an wrapper type.
type wrapper map[string]interface{}

// readID reads "id" from request URL and returns it and nil.
// If there is error it returns 0 and error.
func (app *application) readID(r *http.Request) (int64, error) {
	parameters := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(parameters.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

// writeJSON marshals data structure to encoded JSON response.
// It returns error if there are any issues, else error is nil.
func (app *application) writeJSON(w http.ResponseWriter, status int, data wrapper, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

// readJSON decodes request Body into corresponding Go type. It triages for any potential errors
// and returns corresponding appropriate errors.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, destination interface{}) error {
	maxBytes := 1000000
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(destination)
	if err != nil {
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var unmarshalTypeError *json.UnmarshalTypeError
		var syntaxError *json.SyntaxError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("incorrect form of JSON, character %d", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("incorrect form of JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("incorrect type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("incorrect JSON type, character %d", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			keyName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown key %s", keyName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("must not be more than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	err = decoder.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("not single JSON")
	}
	return nil
}

// readString is helper method on *application that returns string value from the URL query
// string, or the provided default value if no key is found.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

// readCSV is helper method on *application that reads string value from the URL query
// string and splits it into slice on the comma character.
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// readInt is helper method on *application that reads string value from the URL query
// string and converts it to integer. If no key is found it returns the provided default value.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}
