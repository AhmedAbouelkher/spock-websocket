package main

import (
	"os"
	"strings"
)

func IsEnvProduction() bool { return !IsEnvLocal() }
func IsEnvLocal() bool      { return strings.ToLower(os.Getenv("ENV")) == "local" }
