package main

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"fetch.trungnng.github.io/internal/assert"
	"fetch.trungnng.github.io/internal/data"
)

func TestProcessReceiptHandler(t *testing.T) {
	app := newTestApplication()

	ts := newTestServer(app.routes())
	defer ts.Close()

	jsonData := `{
		"retailer": "Target",
		"purchaseDate": "2022-01-01",
		"purchaseTime": "13:01",
		"items": [
		  {
			"shortDescription": "Mountain Dew 12PK",
			"price": "6.49"
		  },{
			"shortDescription": "Emils Cheese Pizza",
			"price": "12.25"
		  },{
			"shortDescription": "Knorr Creamy Chicken",
			"price": "1.26"
		  },{
			"shortDescription": "Doritos Nacho Cheese",
			"price": "3.35"
		  },{
			"shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
			"price": "12.00"
		  }
		],
		"total": "35.35"
	  }`
	myReader := strings.NewReader(jsonData)

	status, _, res := ts.post(t, "/receipts/process", myReader)
	assert.Equal(t, status, http.StatusOK)
	assert.Contains(t, res, "id")
}

func TestGetPointsHandler(t *testing.T) {
	receipt := data.Receipt{
		ID:           "7fb1377b-b223-49d9-a31a-5a02701dd310",
		Retailer:     "Target",
		PurchaseDate: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		PurchaseTime: time.Date(1, time.January, 1, 13, 1, 0, 0, time.UTC),
		Items: []*data.Item{
			{ShortDescription: "Mountain Dew 12PK", Price: 6.49},
			{ShortDescription: "Emils Cheese Pizza", Price: 12.25},
			{ShortDescription: "Knorr Creamy Chicken", Price: 1.26},
			{ShortDescription: "Doritos Nacho Cheese", Price: 3.35},
			{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: 12.00},
		},
		Total: 35.35,
	}

	app := newTestApplication()

	ts := newTestServer(app.routes())
	defer ts.Close()

	app.model = data.NewModels()
	err := app.model.Receipts.Insert(&receipt)
	if err != nil {
		t.Fatal("Unable to insert receipt")
	}

	status, _, res := ts.get(t, "/receipts/7fb1377b-b223-49d9-a31a-5a02701dd310/points")
	assert.Equal(t, status, 200)
	assert.Contains(t, res, "28")
}
