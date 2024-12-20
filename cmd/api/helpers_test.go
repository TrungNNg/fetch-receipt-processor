package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fetch.trungnng.github.io/internal/assert"
	"fetch.trungnng.github.io/internal/data"
)

func TestWriteJSON(t *testing.T) {
	app := newTestApplication()

	ts := newTestServer(app.routes())
	defer ts.Close()

	rr := httptest.NewRecorder()

	headers := http.Header{}
	headers.Set("X-Custom-Header", "CustomValue")

	data := envelope{"message": "Hello, World!"}

	// Call the writeJSON method.
	err := app.writeJSON(rr, http.StatusOK, data, headers)

	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	assert.Equal(t, "CustomValue", rr.Header().Get("X-Custom-Header"))

	// Get the http.Response generated by writeJSON.
	rs := rr.Result()

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, strings.TrimSpace(string(body)), strings.TrimSpace(string(jsonBytes)))
}

func TestReadJSON(t *testing.T) {
	type TestStruct struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
	}

	tests := []struct {
		name           string
		body           string
		expectedErr    string
		shouldSucceed  bool
		expectedStruct TestStruct
	}{
		{
			name:          "Valid JSON",
			body:          `{"name": "test", "value": 123.45}`,
			expectedErr:   "",
			shouldSucceed: true,
			expectedStruct: TestStruct{
				Name:  "test",
				Value: 123.45,
			},
		},
		{
			name:          "Malformed JSON",
			body:          `{"name": "test", "value": 123.45`,
			expectedErr:   "body contains badly-formed JSON",
			shouldSucceed: false,
		},
		{
			name:          "Unknown Field",
			body:          `{"name": "test", "value": 123.45, "extra": "field"}`,
			expectedErr:   "body contains unknown key \"extra\"",
			shouldSucceed: false,
		},
		{
			name:          "Empty Body",
			body:          ``,
			expectedErr:   "body must not be empty",
			shouldSucceed: false,
		},
		{
			name:          "Exceeds Max Size",
			body:          `{"name": "` + string(make([]byte, 1_048_577)) + `"}`,
			expectedErr:   "body must not be larger than 1048576 bytes",
			shouldSucceed: false,
		},
		{
			name:          "Extra JSON Value",
			body:          `{"name": "test", "value": 123.45}{"extra": "value"}`,
			expectedErr:   "body must only contain a single JSON value",
			shouldSucceed: false,
		},
	}

	app := newTestApplication()

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tc.body))
		w := httptest.NewRecorder()

		var result TestStruct
		err := app.readJSON(w, req, &result)
		if tc.shouldSucceed {
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStruct, result)
		} else if tc.expectedErr != "" {
			assert.Contains(t, err.Error(), tc.expectedErr)
		}
	}
}

func TestCalculatePoints(t *testing.T) {
	// Create a test app with a reciept model dependency.
	app := newTestApplication()
	app.model = data.NewModels()

	// Use the model to create mock recipts.
	receipt1 := app.model.Receipts.NewReceipt()
	receipt1.Retailer = "Retailer123"
	receipt1.Total = 100.00
	receipt1.PurchaseDate = time.Date(2024, 12, 13, 0, 0, 0, 0, time.UTC)  // Even day
	receipt1.PurchaseTime = time.Date(2024, 12, 13, 15, 0, 0, 0, time.UTC) // 3:00 PM
	receipt1.Items = []*data.Item{
		{
			ShortDescription: "Milk",
			Price:            2.50,
		},
		{
			ShortDescription: "Bread",
			Price:            1.50,
		},
	}

	receipt2 := app.model.Receipts.NewReceipt()
	receipt2.Retailer = "Shop#42"
	receipt2.Total = 99.99
	receipt2.PurchaseDate = time.Date(2024, 12, 13, 0, 0, 0, 0, time.UTC)  // Even day
	receipt2.PurchaseTime = time.Date(2024, 12, 13, 16, 0, 0, 0, time.UTC) // 4:00 PM
	receipt2.Items = []*data.Item{
		{
			ShortDescription: "Juice",
			Price:            3.75,
		},
	}

	tests := []struct {
		name     string
		receipt  *data.Receipt
		expected int64
	}{
		{
			name:     "Receipt with round dollar amount, multiple of 0.25, and bonus points",
			receipt:  receipt1,
			expected: 50 + 25 + 10 + 6 + 12 + 100, // Round + Multiple + Time + Day + Alphanumeric + Item Points
		},
		{
			name:     "Receipt with non-round total and fewer bonuses",
			receipt:  receipt2,
			expected: 9 + 0, // Alphanumeric only
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := app.calculatePoints(tc.receipt)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestReadIDParam(t *testing.T) {
	app := newTestApplication()

	req := httptest.NewRequest(http.MethodGet, "/receipts/7fb1377b-b223-49d9-a31a-5a02701dd310/points", nil)

	id, err := app.readIDParam(req)
	if err != nil {
		t.Error("Id should be valid.")
	}
	assert.Equal(t, id, "7fb1377b-b223-49d9-a31a-5a02701dd310")
}
