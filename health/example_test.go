package health

import "fmt"

func ExampleHealthState_IsReady() {
	fmt.Println(StatusHealthy.IsReady())
	fmt.Println(StatusDegraded.IsReady())
	fmt.Println(StatusUnhealthy.IsReady())

	// Output:
	// true
	// true
	// false
}

func ExampleHealthState_IsKnown() {
	fmt.Println(StatusHealthy.IsKnown())
	fmt.Println(StatusUnhealthy.IsKnown())
	fmt.Println(HealthState("starting").IsKnown())

	// Output:
	// true
	// true
	// false
}

func ExampleHealthStatus() {
	status := HealthStatus{
		Status:  StatusDegraded,
		Message: "cache disabled",
		Details: map[string]any{"enabled": false},
	}

	fmt.Println(status.Status)
	fmt.Println(status.Message)
	fmt.Println(status.Details["enabled"])

	// Output:
	// degraded
	// cache disabled
	// false
}

func ExampleComponentHealth() {
	component := ComponentHealth{
		HealthStatus: HealthStatus{Status: StatusHealthy},
		Enabled:      true,
	}

	fmt.Println(component.Enabled)
	fmt.Println(component.Status.IsReady())

	// Output:
	// true
	// true
}

func ExampleReadinessStatus() {
	readiness := ReadinessStatus{
		Ready:      false,
		Reason:     "cache offline",
		Components: map[string]bool{"api": true, "cache": false},
	}

	fmt.Println(readiness.Ready)
	fmt.Println(readiness.Reason)
	fmt.Println(readiness.Components["api"])
	fmt.Println(readiness.Components["cache"])

	// Output:
	// false
	// cache offline
	// true
	// false
}
