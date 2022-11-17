package main

import (
	"fmt"
	"testing"
)

func TestCollectorConfigInit(t *testing.T) {
	CollectorConfigInit()

	fmt.Printf("collector config is :\n %v \n", collectorConfig)
}
