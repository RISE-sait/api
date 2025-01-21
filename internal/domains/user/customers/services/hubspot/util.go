package hubspot

import (
	"api/internal/types"
	"api/internal/utils"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func ExecuteHubSpotRequest[T any](s *HubSpotService, method, url string, body any) (*T, *types.HTTPError) {
	var requestBody []byte
	var err error

	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, utils.CreateHTTPError(err.Error(), http.StatusBadRequest)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}

	req.Header.Set("Authorization", "Bearer "+s.ApiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.Client.Do(req)

	if err != nil {
		return nil, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	statusCode := resp.StatusCode
	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}
	if statusCode >= 400 {
		return nil, utils.CreateHTTPError(string(bytes), statusCode)
	}

	var result T

	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}

	// success

	return &result, nil
}
