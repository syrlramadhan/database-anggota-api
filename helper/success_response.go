package helper

import (
	"encoding/json"
	"net/http"

	"github.com/syrlramadhan/database-anggota-api/dto"
)

// WriteJSONSuccess digunakan untuk mengirim response sukses dalam format JSON
func WriteJSONSuccess(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := dto.ListResponseOK{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Data:    data,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
