package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var (
	Policies []Policy
)

type (
	File struct {
		Policies []Policy `yaml:"policies"`
	}
	Policy struct {
		Name             string   `yaml:"name" validate:"required"`
		Query            string   `yaml:"query" validate:"required"`
		App              string   `yaml:"app" validate:"required"`
		CoolDownSec      *int     `yaml:"cooldown_sec"`
		QueryIntervalSec *int     `yaml:"query_interval_sec"`
		Min              *int     `yaml:"min"`
		Max              *int     `yaml:"max"`
		ProtectedIds     []string `yaml:"protected_ids"`
		Regions          []string `yaml:"regions" validate:"required""`
		IncreaseBy       *int     `yaml:"increase_by"`
		DecreaseBy       *int     `yaml:"decrease_by"`
	}
)

var (
	DefaultCoolDownSec      = 30
	DefaultQueryIntervalSec = 30
	DefaultMin              = 0
	DefaultMax              = 0
	DefaultIncreaseBy       = 1
	DefaultDecreaseBy       = 1
)

func LoadConfig(f []byte) error {
	fConfig := File{}
	err := yaml.Unmarshal(f, &fConfig)
	validate := validator.New()
	err = validate.Struct(fConfig)
	if err != nil {
		return fmt.Errorf("error in validation: %w", err)
	}
	//validationErrors := err.(validator.ValidationErrors)
	Policies = fConfig.Policies
	return nil
}
