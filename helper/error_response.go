package helper

import (
	"encoding/json"
	"net/http"

	"github.com/syrlramadhan/database-anggota-api/dto"
)

// WriteJSONError untuk mengirim response error dengan format JSON
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.ListResponseError{
		Code:    statusCode,
		Status:  http.StatusText(statusCode),
		Message: message,
	})
}