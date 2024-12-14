package data

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidReceiptRetailerFormat     = errors.New("invalid retailer format")
	ErrInvalidReceiptPurchaseDateFormat = errors.New("invalid purchaseDate format")
	ErrInvalidReceiptPurchaseTimeFormat = errors.New("invalid purchaseTime format")
	ErrInvalidReceiptAmountFormat       = errors.New("invalid amount format")
)

// Regex patterns to validate JSON inputs format
var (
	retailerRX         = regexp.MustCompile(`^[\w\s\-&]+$`)
	shortDescriptionRX = regexp.MustCompile(`^[\w\s\-]+$`)

	// YYYY-MM-DD
	dateRX = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`)

	// HH:MM in 24-hour format
	timeRX = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d$`)

	// Use for both total and price fields
	amountRX = regexp.MustCompile(`^\d+\.\d{2}$`)
)

type ReceiptRetailer string
type ReceiptShortDescription string
type ReceiptPurchaseDate time.Time
type ReceiptPurchaseTime time.Time
type ReceiptAmount float32

// Custom decoder for the retailer field of the input JSON.
func (rr *ReceiptRetailer) UnmarshalJSON(jsonValue []byte) error {
	// The incomming JSON value will be a string in double quote, so we need to
	// remove the double quote
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidReceiptRetailerFormat
	}

	// Remove leading and trailing whitespaces.
	unquotedJSONValue = strings.TrimSpace(unquotedJSONValue)

	// Validate that the string matches the required pattern.
	if !retailerRX.MatchString(unquotedJSONValue) {
		return ErrInvalidReceiptRetailerFormat
	}

	// Assign the trimmed value to the ReceiptString type.
	*rr = ReceiptRetailer(strings.TrimSpace(unquotedJSONValue))
	return nil
}

// shortDescription field
func (rs *ReceiptShortDescription) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidReceiptRetailerFormat
	}

	unquotedJSONValue = strings.TrimSpace(unquotedJSONValue)

	if !shortDescriptionRX.MatchString(unquotedJSONValue) {
		return ErrInvalidReceiptRetailerFormat
	}

	*rs = ReceiptShortDescription(strings.TrimSpace(unquotedJSONValue))
	return nil
}

// purchaseDate field
func (rd *ReceiptPurchaseDate) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidReceiptPurchaseDateFormat
	}

	unquotedJSONValue = strings.TrimSpace(unquotedJSONValue)

	if !dateRX.MatchString(unquotedJSONValue) {
		return ErrInvalidReceiptPurchaseDateFormat
	}

	// In the absence of a time zone indicator, Parse returns a time in UTC.
	parsedTime, err := time.Parse("2006-01-02", unquotedJSONValue)
	if err != nil {
		return ErrInvalidReceiptPurchaseDateFormat
	}

	*rd = ReceiptPurchaseDate(parsedTime)
	return nil
}

// purchaseTime field
func (rt *ReceiptPurchaseTime) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidReceiptPurchaseTimeFormat
	}

	unquotedJSONValue = strings.TrimSpace(unquotedJSONValue)

	if !timeRX.MatchString(unquotedJSONValue) {
		return ErrInvalidReceiptPurchaseTimeFormat
	}

	parsedTime, err := time.Parse("15:04", unquotedJSONValue)
	if err != nil {
		return ErrInvalidReceiptPurchaseTimeFormat
	}
	*rt = ReceiptPurchaseTime(parsedTime)
	return nil
}

// total and price fields
func (ra *ReceiptAmount) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return err
	}

	unquotedJSONValue = strings.TrimSpace(unquotedJSONValue)

	if !amountRX.MatchString(unquotedJSONValue) {
		return ErrInvalidReceiptAmountFormat
	}

	amount, err := strconv.ParseFloat(unquotedJSONValue, 32)
	if err != nil {
		return ErrInvalidReceiptAmountFormat
	}

	*ra = ReceiptAmount(float32(amount))
	return nil
}
