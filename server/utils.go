package main

import "os"

func GetenvDef(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}
