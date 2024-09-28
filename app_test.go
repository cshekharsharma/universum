package main

import (
	"fmt"
	"os"
	"testing"
	"universum/config"
)

func TestMain(m *testing.M) {
	fmt.Println("Setup resources for tests")
	config.Store = config.GetSkeleton()
	validator := config.NewConfigValidator()
	_, _ = validator.Validate(config.Store)

	code := m.Run()

	fmt.Println("Teardown resources after tests")

	os.Exit(code)
}
