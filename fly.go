package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/danthegoodman1/FlyMachineAutoscaler/utils"
	"io"
	"net/http"
	"time"
)

type FlyListMachinesResp struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	State    string `json:"state"`
	Region   string `json:"region"`
	ImageRef struct {
		Registry   string         `json:"registry"`
		Repository string         `json:"repository"`
		Tag        string         `json:"tag"`
		Digest     string         `json:"digest"`
		Labels     map[string]any `json:"labels"`
	} `json:"image_ref"`
	InstanceId string    `json:"instance_id"`
	Version    string    `json:"version"`
	PrivateIp  string    `json:"private_ip"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Config     struct {
		Env  interface{} `json:"env"`
		Init struct {
			Exec       interface{} `json:"exec"`
			Entrypoint interface{} `json:"entrypoint"`
			Cmd        interface{} `json:"cmd"`
			Tty        bool        `json:"tty"`
		} `json:"init"`
		Image    string      `json:"image"`
		Metadata interface{} `json:"metadata"`
		Restart  struct {
			Policy string `json:"policy"`
		} `json:"restart"`
		Guest struct {
			CpuKind  string `json:"cpu_kind"`
			Cpus     int    `json:"cpus"`
			MemoryMb int    `json:"memory_mb"`
		} `json:"guest"`
		Metrics interface{} `json:"metrics"`
	} `json:"config"`
	Events []struct {
		Type      string `json:"type"`
		Status    string `json:"status"`
		Source    string `json:"source"`
		Timestamp int64  `json:"timestamp"`
		Request   struct {
			ExitEvent struct {
				ExitCode      int  `json:"exit_code"`
				GuestExitCode int  `json:"guest_exit_code"`
				GuestSignal   int  `json:"guest_signal"`
				OomKilled     bool `json:"oom_killed"`
				RequestedStop bool `json:"requested_stop"`
				Restarting    bool `json:"restarting"`
				Signal        int  `json:"signal"`
			} `json:"exit_event"`
			RestartCount int `json:"restart_count"`
		} `json:"request,omitempty"`
	} `json:"events"`
	LeaseNonce string `json:"LeaseNonce"`
}

func ListMachines(ctx context.Context, app string) (*[]FlyListMachinesResp, error) {
	return flyRequest[[]FlyListMachinesResp](ctx, fmt.Sprintf("/%s/machines", app), "GET", nil)
}

func CreateMachine(ctx context.Context, app string, config any) error {
	return nil
}

func DeleteMachine(ctx context.Context, app, instanceID string) error {
	return nil
}

type FlyMachineConfig struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	Region     string `json:"region"`
	InstanceId string `json:"instance_id"`
	PrivateIp  string `json:"private_ip"`
	Config     struct {
		Init struct {
		} `json:"init"`
		Services []struct {
			Protocol     string `json:"protocol"`
			InternalPort int    `json:"internal_port"`
			Ports        []struct {
				Port int `json:"port"`
			} `json:"ports"`
			ForceInstanceKey interface{} `json:"force_instance_key"`
		} `json:"services"`
		Checks map[string]struct {
			Port     int    `json:"port"`
			Type     string `json:"type"`
			Interval string `json:"interval"`
			Timeout  string `json:"timeout"`
		} `json:"checks"`
		Image   string `json:"image"`
		Restart struct {
		} `json:"restart"`
		Guest struct {
			CpuKind  string `json:"cpu_kind"`
			Cpus     int    `json:"cpus"`
			MemoryMb int    `json:"memory_mb"`
		} `json:"guest"`
		Mounts []struct {
			Path   string `json:"path"`
			Volume string `json:"volume"`
		} `json:"mounts"`
	} `json:"config"`
	ImageRef struct {
		Registry   string `json:"registry"`
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
		Digest     string `json:"digest"`
		Labels     struct {
			Maintainer string `json:"maintainer"`
		} `json:"labels"`
	} `json:"image_ref"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Events    []struct {
		Id        string `json:"id"`
		Type      string `json:"type"`
		Status    string `json:"status"`
		Source    string `json:"source"`
		Timestamp int64  `json:"timestamp"`
	} `json:"events"`
	Checks []struct {
		Name      string    `json:"name"`
		Status    string    `json:"status"`
		Output    string    `json:"output"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"checks"`
}

func GetMachineInfo(ctx context.Context, app, instanceID string) (*FlyMachineConfig, error) {
	return flyRequest[FlyMachineConfig](ctx, fmt.Sprintf("/%s/machines/%s", app, instanceID), "GET", nil)
}

func flyRequest[T any](ctx context.Context, path, method string, body any) (*T, error) {

	bodyB, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.machines.dev/v1/apps"+path, bytes.NewReader(bodyB))
	if err != nil {
		return nil, fmt.Errorf("error creating new request : %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+utils.Env_FlyAPIToken)

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

	var resBody T
	err = json.Unmarshal(resB, &resBody)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response: %w", err)
	}
	return &resBody, nil
}
