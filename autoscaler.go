package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/danthegoodman1/FlyMachineAutoscaler/utils"
	"io"
	"net/http"
	"time"
)

var (
	ErrHighStatusCode = errors.New("high status code")
)

type ()

type PromQueryResponse struct {
	Status    string `json:"status"`
	IsPartial bool   `json:"isPartial"`
	Data      struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]any `json:"metric"`
			Value  []interface{}  `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func QueryPrometheus(ctx context.Context, query string, t time.Time) (*PromQueryResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.fly.io/prometheus/%s/api/v1/query", utils.Env_FlyOrg), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request : %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+utils.Env_FlyAPIToken)

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("time", fmt.Sprint(t.Unix()))
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error in http.Do: %w", err)
	}
	defer res.Body.Close()
	resB, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error in io.ReadAll: %w", err)
	}

	if res.StatusCode > 299 {
		return nil, fmt.Errorf("high status code %d --- %s --- %w", res.StatusCode, string(resB), ErrHighStatusCode)
	}

	logger.Debug().Str("query", query).Str("responseBody", string(resB)).Msg("got query response")
	fmt.Println(string(resB))

	var resBody PromQueryResponse
	err = json.Unmarshal(resB, &resBody)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response: %w", err)
	}

	return &resBody, nil
}
