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
	ErrVolumesNotSupported      = errors.New("volumes not suppported")

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
policyLoop:
	for _, policy := range config.Policies {
		// First we check for mins and maxes
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		machines, err := ListMachines(ctx, policy.App)
		if err != nil {
			return fmt.Errorf("error in ListMachines: %w", err)
		}

		// Handle regions
		// Group by region
		regionMap := map[string][]FlyListMachinesResp{}
		for _, machine := range *machines {
			if _, exists := regionMap[machine.Region]; !exists {
				regionMap[machine.Region] = []FlyListMachinesResp{}
			}
			regionMap[machine.Region] = append(regionMap[machine.Region], machine)
		}
		// Check for min and max
		var scaleUpRegions, scaleDownRegions []string
		for region, regionMachines := range regionMap {
			if policy.Min != nil && *policy.Min > len(regionMachines) {
				scaleUpRegions = append(scaleUpRegions, region)
			}
			if policy.Max != nil && *policy.Max < len(regionMachines) {
				scaleDownRegions = append(scaleDownRegions, region)
			}
		}

		// Get the metrics
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*30)
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

		// For anything above, we need to scale up
		for _, aboveRegion := range aboveRegions {
			// Verify we are able to scale up based on cooldowns
			if _, exists := aboveRegion.Metric["region"]; !exists {
				logger.Err(ErrQueryResultMissingRegion).Str("policy", policy.Name).Msg("policy query result missing region, skipping policy")
				continue policyLoop
			}
			region := aboveRegion.Metric["region"].(string)
			err = scaleUpRegion(ctx, policy, region, (*machines)[0])
			if err != nil {
				return fmt.Errorf("error in scaleUpRegion: %w", err)
			}
		}

		// For anything below, we need to scale down
		for _, belowRegion := range belowRegions {
			// Verify we are able to scale up based on cooldowns
			if _, exists := belowRegion.Metric["region"]; !exists {
				logger.Err(ErrQueryResultMissingRegion).Str("policy", policy.Name).Msg("policy query result missing region, skipping policy")
				continue policyLoop
			}
			region := belowRegion.Metric["region"].(string)
			err = scaleDownRegion(ctx, policy, region, regionMap[region])
			if err != nil {
				return fmt.Errorf("error in scaleDownRegion: %w", err)
			}
		}
	}
	return nil
}

func scaleUpRegion(ctx context.Context, policy config.Policy, region string, machineToClone FlyListMachinesResp) error {
	key := policy.App + region
	checkTime := time.Now()
	if lastUp, exists := LastScaleUp[key]; !exists || checkTime.Sub(lastUp) > time.Duration(utils.Deref(policy.CoolDownSec, config.DefaultCoolDownSec))*time.Second {
		// TODO: check for volume, get volume config, and create one
		cloneConfig, err := GetMachineInfo(ctx, policy.App, machineToClone.Id)
		if err != nil {
			return fmt.Errorf("error in GetMachineInfo: %w", err)
		}
		if len(cloneConfig.Config.Mounts) > 0 {
			return fmt.Errorf("machine %s to clone had a volume: %w", machineToClone.Id, ErrVolumesNotSupported)
		}
		for i := 0; i < utils.Deref(policy.IncreaseBy, config.DefaultIncreaseBy); i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			err := CreateMachine(ctx, policy.App, cloneConfig)
			if err != nil {
				return fmt.Errorf("error in CreateMachine: %w", err)
			}
		}
		LastScaleUp[key] = checkTime
	}
	return nil
}

func scaleDownRegion(ctx context.Context, policy config.Policy, region string, regionalMachines []FlyListMachinesResp) error {
	randSource := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	key := policy.App + region
	checkTime := time.Now()
	if lastUp, exists := LastScaleUp[key]; !exists || checkTime.Sub(lastUp) > time.Duration(utils.Deref(policy.CoolDownSec, config.DefaultCoolDownSec))*time.Second {
		candidateMachines := lo.Filter(regionalMachines, func(item FlyListMachinesResp, index int) bool {
			return !lo.Contains(policy.ProtectedIds, item.Id)
		})
		randMachine := randSource.Intn(len(candidateMachines))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		machine, err := GetMachineInfo(ctx, policy.App, candidateMachines[randMachine].Id)
		if err != nil {
			return fmt.Errorf("error in GetMachineInfo: %w", err)
		}
		err = DeleteMachine(ctx, policy.App, candidateMachines[randMachine].Id)
		if err != nil {
			return fmt.Errorf("error in DeleteMachine: %w", err)
		}
		if !policy.SaveVolumes && len(machine.Config.Mounts) > 0 {
			// TODO: If there was a volume on the machine, destroy that too
			logger.Error().Err(ErrVolumesNotSupported).Str("machineID", machine.Id).Interface("mounts", machine.Config.Mounts).Msg("volume found on machine, ignoring on scale down!")
		}
		LastScaleDown[key] = checkTime
	}
	return nil
}
