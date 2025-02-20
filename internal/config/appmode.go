package config

import (
	"log"
	"os"
	"strings"
)

type AppMode string

const (
	Development = AppMode("development")
	Staging     = AppMode("staging")
	Production  = AppMode("production")
)

func NewFromEnv(key string) AppMode {
	modeStr := os.Getenv(key)
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
