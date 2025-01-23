package response_handlers

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, err *errLib.CommonError) {

	log.Println("RespondWithError")
	w.WriteHeader(err.HTTPCode)
	response := map[string]interface{}{
		"error": map[string]string{
			"message": err.Message,
		},
	}
	log.Println("after  RespondWithError")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}
