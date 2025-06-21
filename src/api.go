package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func SendOpenAIRequest(body OpenAIAPIRequest, apiKey string) (GPTResponse, error) {
	gptResponse := GPTResponse{}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		return gptResponse, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(bodyBytes),
	)

	if err != nil {
		return gptResponse, fmt.Errorf("Error creating request: %+v\n", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return gptResponse, fmt.Errorf("Error making request: %+v\n", err)
	}

	fmt.Printf("StatusCode: %d\n", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return gptResponse, fmt.Errorf("Error reading response: %+v\n", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err = json.Unmarshal(respBody, &gptResponse)

		if err != nil {
			return gptResponse, fmt.Errorf(
				"Error unmarshalling response into GPTResponse: %+v\n",
				err,
			)
		}

		return gptResponse, nil
	}

	response := map[string]any{}

	err = json.Unmarshal(respBody, &response)

	if err != nil {
		return gptResponse, fmt.Errorf("Error unmarshalling response to map: %+v\n", err)
	}

	return gptResponse, fmt.Errorf("Welp")
}
