package main

import (
	"fmt"
	"net/http"
	"time"

	"fetch.trungnng.github.io/internal/data"
	"fetch.trungnng.github.io/internal/validator"
)

// Submits a receipt for processing
func (app *application) processReceiptHandler(w http.ResponseWriter, r *http.Request) {
	errorMessage := "The receipt is invalid"

	// The Price and Total fields are defined as pointers to allow distinguishing between
	// missing values and zero values. This is important because a value of 0.0 is a valid
	// price or total (e.g., in cases of price discounts or user have extra credits).
	var input struct {
		Retailer     data.ReceiptRetailer     `json:"retailer"`
		PurchaseDate data.ReceiptPurchaseDate `json:"purchaseDate"`
		PurchaseTime data.ReceiptPurchaseTime `json:"purchaseTime"`
		Items        []struct {
			ShortDescription data.ReceiptShortDescription `json:"shortDescription"`
			Price            *data.ReceiptAmount          `json:"price"`
		} `json:"items"`
		Total *data.ReceiptAmount `json:"total"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		fmt.Println(err.Error())
		app.badRequestResponse(w, r, errorMessage)
		return
	}

	// Check if the receipt has a total field.
	if input.Total == nil {
		app.badRequestResponse(w, r, errorMessage)
		return
	}

	// Check if all items in the receipt have a price field.
	for _, item := range input.Items {
		if item.Price == nil {
			app.badRequestResponse(w, r, errorMessage)
			return
		}
	}

	// Copy the values from the input struct to a new Receipt struct.
	receipt := app.model.Receipts.NewReceipt()
	receipt.Retailer = string(input.Retailer)
	receipt.PurchaseDate = time.Time(input.PurchaseDate)
	receipt.PurchaseTime = time.Time(input.PurchaseTime)
	for _, item := range input.Items {
		i := app.model.Receipts.NewReceiptItem()
		i.ShortDescription = string(item.ShortDescription)
		i.Price = float32(*item.Price)
		receipt.Items = append(receipt.Items, i)
	}
	receipt.Total = float32(*input.Total)

	// Validate reciept data
	v := validator.New()

	if data.ValidateReceipt(v, receipt); !v.Valid() {
		app.badRequestResponse(w, r, errorMessage)
		return
	}

	// Save to DB
	err = app.model.Receipts.Insert(receipt)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Response to client
	err = app.writeJSON(w, 200, envelope{"id": receipt.ID}, nil)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

// getPointsHandler handles the HTTP request to retrieve the points for a specific receipt by ID.
// 1. Extracts the `id` from the URL path parameters.
// 2. Attempts to retrieve the receipt from the database using the provided `id`.
// 3. Calculates the total points for the retrieved receipt.
// 4. Responds with the calculated points in JSON format.
func (app *application) getPointsHandler(w http.ResponseWriter, r *http.Request) {
	errorMessage := "No receipt found for that id"

	// Extract the receipt ID from the URL path parameters.
	id, err := app.readIDParam(r)
	if err != nil {
		app.receiptIDNotFoundResponse(w, r, errorMessage)
		return
	}

	// Check if the ID is empty.
	if id == "" {
		app.receiptIDNotFoundResponse(w, r, errorMessage)
		return
	}

	// Retrieve the receipt from the database by ID.
	receipt, err := app.model.Receipts.Get(id)
	if err != nil {
		app.receiptIDNotFoundResponse(w, r, errorMessage)
		return
	}

	// Calculate the points for the retrieved receipt.
	points := app.calculatePoints(receipt)

	// Send the response with the calculated points in JSON format.
	err = app.writeJSON(w, http.StatusOK, envelope{"points": points}, nil)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
