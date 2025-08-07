package util

import (
	"encoding/json"
	"net/http"
)

func ReadFromRequestBody(request *http.Request, result interface{}) error {
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(result)
	return err
}

// func WriteToResponseBody(writer http.ResponseWriter, response interface{}) error {
// 	writer.Header().Add("Content-Type", "application/json")
// 	encoder := json.NewEncoder(writer)
// 	err := encoder.Encode(response)
// 	return err
// }
