package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"

	"fetch.trungnng.github.io/internal/data"
	"github.com/julienschmidt/httprouter"
)

// Envelope type for JSON response.
type envelope map[string]any

// Retrieve the "id" URL parameter from the current request context.
// Return empty string if no id param found.
func (app *application) readIDParam(r *http.Request) (string, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id := params.ByName("id")

	idRX := regexp.MustCompile(`^\S+$`)
	if !idRX.MatchString(id) {
		return "", errors.New("invalid id param pattern")
	}

	return id, nil
}

// writeJSON sends a JSON response to the client.
//
// Parameters:
// - w: The http.ResponseWriter where the response will be written.
// - status: The HTTP status code for the response.
// - data: The data to encode into the JSON response body.
// - headers: A map of additional HTTP headers to include in the response.
//
// Returns:
// - An error if encoding the data to JSON fails or writing to the ResponseWriter encounters an issue.
func (app *application) writeJSON(
	w http.ResponseWriter,
	status int,
	data envelope,
	headers http.Header,
) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Add response's headers
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// readJSON reads and decodes a JSON request body into the specified destination `dst`.
// It includes validation to ensure the request body adheres to expected size and format constraints.
//
// Parameters:
// - w: The http.ResponseWriter for sending error responses if necessary.
// - r: The *http.Request containing the JSON body to decode.
// - dst: A pointer to the destination object where the JSON will be unmarshaled.
//
// Returns:
// - An error describing why the JSON decoding failed.
//
// Validates scenarios:
//   - Malformed JSON.
//   - Unexpected data types in the JSON.
//   - Unknown JSON fields.
//   - Empty request bodies.
//   - Oversized bodies
//   - Multiple JSON values.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize json.Decoder.
	dec := json.NewDecoder(r.Body)

	// Decode() will now return error if JSON has unknown fields.
	dec.DisallowUnknownFields()

	// Decode the request body to the destination.
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Call Decode() again using struct{}{} which will consume extra JSON data then discard it.
	// If the request body only contained a single JSON value this will
	// return an io.EOF error. If we get any other error or nil then there is extra data
	// so we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// calculatePoints calculates the total points for a given receipt
func (app *application) calculatePoints(receipt *data.Receipt) int64 {
	var totalPoints int64

	// One point for every alphanumeric character in the retailer name.
	retailer := strings.TrimSpace(receipt.Retailer)
	for _, char := range retailer {
		if isAlphanumeric(char) {
			totalPoints++
		}
	}

	// 50 points if the total is a round dollar amount with no cents.
	if receipt.Total == float32(int(receipt.Total)) {
		totalPoints += 50
	}

	// 25 points if the total is a multiple of 0.25.
	if math.Mod(float64(receipt.Total), 0.25) == 0 {
		totalPoints += 25
	}

	// 5 points for every two items on the receipt.
	totalPoints += int64((len(receipt.Items) / 2) * 5)

	// Points for items with descriptions that are multiples of 3 in length.
	for _, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc)%3 == 0 {
			points := math.Ceil(float64(item.Price) * 0.2)
			totalPoints += int64(points)
		}
	}

	// 6 points if the day in the purchase date is odd.
	if receipt.PurchaseDate.Day()%2 != 0 {
		totalPoints += 6
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	hour, _, _ := receipt.PurchaseTime.Clock()
	if hour >= 14 && hour < 16 {
		totalPoints += 10
	}

	return totalPoints
}

// isAlphanumeric checks if a rune is an alphanumeric character.
func isAlphanumeric(char rune) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || ('0' <= char && char <= '9')
}
