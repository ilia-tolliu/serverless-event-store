package config

import (
	"log"
	"os"
	"strings"
)

type AppMode string

const appModeKey = "EVENT_STORE_MODE"

const (
	Development = AppMode("development")
	Staging     = AppMode("staging")
	Production  = AppMode("production")
)

func NewFromEnv() AppMode {
	modeStr := os.Getenv(appModeKey)
	var mode AppMode

	switch strings.ToLower(modeStr) {
	case string(Development):
		mode = Development
	case string(Staging):
		mode = Staging
	case string(Production):
		mode = Production
	default:
		log.Fatalf("unknown app mode: [%s]", modeStr)
	}

	return mode
}

func (mode AppMode) IsDevelopment() bool {
	return mode == Development
}

func (mode AppMode) IsStaging() bool {
	return mode == Staging
}

func (mode AppMode) IsProduction() bool {
	return mode == Production
}

func (mode AppMode) String() string {
	return string(mode)
}
