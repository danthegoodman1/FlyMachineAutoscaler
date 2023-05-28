package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/danthegoodman1/FlyMachineAutoscaler/config"
	"github.com/danthegoodman1/FlyMachineAutoscaler/utils"
	"github.com/samber/lo"
	"io"
	"math/rand"
	"net/http"
	"time"
)

var (
	ErrHighStatusCode           = errors.New("high status code")
	ErrQueryResultMissingRegion = errors.New("query result missing region")

	LastScaleUp   map[string]time.Time
	LastScaleDown map[string]time.Time
)

type PromQueryResponse struct {
	Status    string `json:"status"`
	IsPartial bool   `json:"isPartial"`
	Data      struct {
		ResultType string   `json:"resultType"`
		Result     []Result `json:"result"`
	} `json:"data"`
}

type Result struct {
	Metric map[string]any `json:"metric"`
	Value  []interface{}  `json:"value"`
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

func tickLoop() error {
	randSource := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
policyLoop:
	for _, policy := range config.Policies {
		// Get the metrics
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		metrics, err := QueryPrometheus(ctx, policy.Query, time.Now())
		if err != nil {
			return fmt.Errorf("error in QueryPrometheus for policy %s: %w", policy.Name, err)
		}

		// Check if the metrics are above or below in any region
		aboveRegions := lo.Filter(metrics.Data.Result, func(item Result, index int) bool {
			return item.Value[0].(float64) > policy.HighThreshold
		})

		belowRegions := lo.Filter(metrics.Data.Result, func(item Result, index int) bool {
			return item.Value[0].(float64) < policy.LowThreshold
		})

		if len(aboveRegions) < 0 && len(belowRegions) < 0 {
			logger.Debug().Msg("nothing needs scaling, ending loop")
			return nil
		}

		// Something needs scaling, let's get the machines
		machines, err := ListMachines(ctx, policy.App)
		if err != nil {
			return fmt.Errorf("error in ListMachines: %w", err)
		}

		// For anything above, we need to scale up
		for _, aboveRegion := range aboveRegions {
			// Verify we are able to scale up based on cooldowns
			if _, exists := aboveRegion.Metric["region"]; !exists {
				logger.Err(ErrQueryResultMissingRegion).Str("policy", policy.Name).Msg("policy query result missing region, skipping policy")
				continue policyLoop
			}
			key := policy.App + aboveRegion.Metric["region"].(string)
			checkTime := time.Now()
			if lastUp, exists := LastScaleUp[key]; !exists || checkTime.Sub(lastUp) > time.Duration(utils.Deref(policy.CoolDownSec, config.DefaultCoolDownSec))*time.Second {
				// TODO: get machine config
				for i := 0; i < utils.Deref(policy.IncreaseBy, config.DefaultIncreaseBy); i++ {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
					defer cancel()
					err := CreateMachine(ctx, policy.App, nil)
					if err != nil {
						return fmt.Errorf("error in CreateMachine: %w", err)
					}
				}
				LastScaleUp[key] = checkTime
			}
		}

		// For anything below, we need to scale down
		for _, belowRegion := range belowRegions {
			// Verify we are able to scale up based on cooldowns
			if _, exists := belowRegion.Metric["region"]; !exists {
				logger.Err(ErrQueryResultMissingRegion).Str("policy", policy.Name).Msg("policy query result missing region, skipping policy")
				continue policyLoop
			}
			key := policy.App + belowRegion.Metric["region"].(string)
			checkTime := time.Now()
			if lastUp, exists := LastScaleUp[key]; !exists || checkTime.Sub(lastUp) > time.Duration(utils.Deref(policy.CoolDownSec, config.DefaultCoolDownSec))*time.Second {
				regionMachines := lo.Filter(*machines, func(item FlyListMachinesResp, index int) bool {
					return item.Region == belowRegion.Metric["region"].(string)
				})
				randMachine := randSource.Intn(len(regionMachines))
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				err = DeleteMachine(ctx, policy.App, regionMachines[randMachine].Id)
				if err != nil {
					return fmt.Errorf("error in DeleteMachine: %w", err)
				}
				LastScaleDown[key] = checkTime
			}
		}
	}
	return nil
}
