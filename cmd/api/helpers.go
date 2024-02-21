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

	"github.com/Light2Dark/greenlight/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// Accepts id param of >= 1
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id paramater")
	}

	return id, nil
}

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// json.NewEncoder(w).Encode(data)

// will decode json to dst
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&dst)
	if err != nil {
		var (
			syntaxError           *json.SyntaxError
			unmarshalTypeError    *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
			maxBytesError         *http.MaxBytesError
		)

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("request body contains badly formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF): // sometimes decode gets this error for bad syntax as well, it's a bug
			return errors.New("request body contains badly formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("request body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("request body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("request body is empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			var fieldName = strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown fieldname %s in request body", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("request body contains more than %v bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("request body contains more than one JSON value")
	}

	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key) // extract value from params of url
	if s == "" {
		return defaultValue
	}

	return s
}

// comma-sep values
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	val := qs.Get(key)

	if val == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return intVal
}
