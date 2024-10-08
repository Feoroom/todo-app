package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"library/internal/validation"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type envelope map[string]interface{}

func (app *Application) readID(params httprouter.Params) (int64, error) {

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 0 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func (app *Application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	resp, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(status)
	w.Write(resp)

	return nil
}

func (app *Application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {

	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalError):
			if unmarshalError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}

	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain a single JSON value")
	}

	return nil
}

func (app *Application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *Application) readCSV(qs url.Values, key string, defaultValues []string) []string {

	csv := qs.Get(key)

	if csv == "" {
		return defaultValues
	}

	return strings.Split(csv, ",")
}

func (app *Application) readInt(qs url.Values, key string, defaultValue int, v *validation.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be integer value")
		return defaultValue
	}

	return i
}

func (app *Application) readDate(qs url.Values, key string, defaultValue time.Time, v *validation.Validator) time.Time {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		v.AddError(key, "must be a valid date in format YYYY-MM-DD")
		return defaultValue
	}

	return t
}

// isPreflight checks if the request is a preflight request.
func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}

func (app *Application) background(fn func()) {
	app.wg.Add(1)

	go func() {

		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(err.(error).Error())
			}
		}()

		fn()
	}()
}
