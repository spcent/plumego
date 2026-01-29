package core

import (
	"reflect"
	"testing"
)

// TestScopedLifecycle tests the scoped lifecycle functionality
func TestScopedLifecycle(t *testing.T) {
	container := NewDIContainer()

	// Register a scoped service
	container.RegisterFactory(
		reflect.TypeOf((*ScopedService)(nil)),
		func(c *DIContainer) any {
			return &ScopedService{ID: "scoped"}
		},
		Scoped,
	)

	// Register a singleton service that depends on scoped
	container.RegisterFactory(
		reflect.TypeOf((*SingletonWithScoped)(nil)),
		func(c *DIContainer) any {
			scoped, err := c.Resolve(reflect.TypeOf((*ScopedService)(nil)))
			if err != nil {
				t.Fatalf("Failed to resolve scoped service: %v", err)
			}
			return &SingletonWithScoped{Scoped: scoped.(*ScopedService)}
		},
		Singleton,
	)

	// Test 1: Same resolution chain should get same scoped instance
	t.Run("SameResolutionChainSameInstance", func(t *testing.T) {
		// Resolve singleton which will resolve scoped within the same chain
		singleton1, err := container.Resolve(reflect.TypeOf((*SingletonWithScoped)(nil)))
		if err != nil {
			t.Fatalf("Failed to resolve singleton: %v", err)
		}

		singleton2, err := container.Resolve(reflect.TypeOf((*SingletonWithScoped)(nil)))
		if err != nil {
			t.Fatalf("Failed to resolve singleton: %v", err)
		}

		// Both singletons should have the same scoped instance
		s1 := singleton1.(*SingletonWithScoped)
		s2 := singleton2.(*SingletonWithScoped)

		if s1.Scoped != s2.Scoped {
			t.Errorf("Expected same scoped instance, got different instances")
		}
	})

	// Test 2: Different resolution chains should get different scoped instances
	t.Run("DifferentResolutionChainsDifferentInstances", func(t *testing.T) {
		// Clear scoped instances to simulate new resolution chains
		container.mu.Lock()
		container.scopedInstances = make(map[uint64]map[reflect.Type]any)
		container.mu.Unlock()

		// Resolve scoped service in first chain
		scoped1, err := container.Resolve(reflect.TypeOf((*ScopedService)(nil)))
		if err != nil {
			t.Fatalf("Failed to resolve scoped service: %v", err)
		}

		// Clear scoped instances to simulate new resolution chain
		container.mu.Lock()
		container.scopedInstances = make(map[uint64]map[reflect.Type]any)
		container.mu.Unlock()

		// Resolve scoped service in second chain
		scoped2, err := container.Resolve(reflect.TypeOf((*ScopedService)(nil)))
		if err != nil {
			t.Fatalf("Failed to resolve scoped service: %v", err)
		}

		// Should be different instances
		if scoped1 == scoped2 {
			t.Errorf("Expected different scoped instances, got same instance")
		}
	})
}

// TestScopedWithDependencies tests scoped services with dependencies
func TestScopedWithDependencies(t *testing.T) {
	container := NewDIContainer()

	// Register a singleton dependency
	container.RegisterFactory(
		reflect.TypeOf((*SingletonDependency)(nil)),
		func(c *DIContainer) any {
			return &SingletonDependency{ID: "singleton"}
		},
		Singleton,
	)

	// Register a scoped service that depends on singleton
	container.RegisterFactory(
		reflect.TypeOf((*ScopedWithDependency)(nil)),
		func(c *DIContainer) any {
			singleton, err := c.Resolve(reflect.TypeOf((*SingletonDependency)(nil)))
			if err != nil {
				t.Fatalf("Failed to resolve singleton dependency: %v", err)
			}
			return &ScopedWithDependency{
				ID:        "scoped",
				Dep:       singleton.(*SingletonDependency),
				CreatedAt: 12345, // Mock timestamp
			}
		},
		Scoped,
	)

	// Test that scoped service gets singleton dependency correctly
	t.Run("ScopedGetsSingletonDependency", func(t *testing.T) {
		scoped1, err := container.Resolve(reflect.TypeOf((*ScopedWithDependency)(nil)))
		if err != nil {
			t.Fatalf("Failed to resolve scoped service: %v", err)
		}

		scoped2, err := container.Resolve(reflect.TypeOf((*ScopedWithDependency)(nil)))
		if err != nil {
			t.Fatalf("Failed to resolve scoped service: %v", err)
		}

		s1 := scoped1.(*ScopedWithDependency)
		s2 := scoped2.(*ScopedWithDependency)

		// Scoped instances should be different
		if s1 == s2 {
			t.Errorf("Expected different scoped instances")
		}

		// But both should have the same singleton dependency
		if s1.Dep != s2.Dep {
			t.Errorf("Expected same singleton dependency in both scoped instances")
		}

		// Verify the dependency is correct
		if s1.Dep.ID != "singleton" {
			t.Errorf("Expected singleton dependency ID 'singleton', got '%s'", s1.Dep.ID)
		}
	})
}

// TestScopedWithStructInjection tests scoped services with struct field injection
func TestScopedWithStructInjection(t *testing.T) {
	container := NewDIContainer()

	// Register a singleton service
	container.RegisterFactory(
		reflect.TypeOf((*SingletonService)(nil)),
		func(c *DIContainer) any {
			return &SingletonService{ID: "singleton"}
		},
		Singleton,
	)

	// Register a scoped service with inject tags
	container.RegisterFactory(
		reflect.TypeOf((*ScopedWithInjection)(nil)),
		func(c *DIContainer) any {
			return &ScopedWithInjection{}
		},
		Scoped,
	)

	// Inject dependencies into the scoped service
	t.Run("ScopedInjection", func(t *testing.T) {
		scoped1 := &ScopedWithInjection{}
		err := container.Inject(scoped1)
		if err != nil {
			t.Fatalf("Failed to inject into scoped service: %v", err)
		}

		scoped2 := &ScopedWithInjection{}
		err = container.Inject(scoped2)
		if err != nil {
			t.Fatalf("Failed to inject into scoped service: %v", err)
		}

		// Both scoped instances should have the same singleton dependency
		if scoped1.Singleton != scoped2.Singleton {
			t.Errorf("Expected same singleton dependency in both scoped instances")
		}

		// But the scoped instances themselves should be different
		if scoped1 == scoped2 {
			t.Errorf("Expected different scoped instances")
		}
	})
}

// TestScopedLifecycleWithTransient tests interaction between scoped and transient lifecycles
func TestScopedLifecycleWithTransient(t *testing.T) {
	container := NewDIContainer()

	// Register a transient service
	container.RegisterFactory(
		reflect.TypeOf((*TransientService)(nil)),
		func(c *DIContainer) any {
			return &TransientService{ID: "transient"}
		},
		Transient,
	)

	// Register a scoped service that depends on transient
	container.RegisterFactory(
		reflect.TypeOf((*ScopedWithTransient)(nil)),
		func(c *DIContainer) any {
			transient, err := c.Resolve(reflect.TypeOf((*TransientService)(nil)))
			if err != nil {
				t.Fatalf("Failed to resolve transient service: %v", err)
			}
			return &ScopedWithTransient{Transient: transient.(*TransientService)}
		},
		Scoped,
	)

	t.Run("ScopedWithTransient", func(t *testing.T) {
		// Resolve scoped service multiple times
		scoped1, err := container.Resolve(reflect.TypeOf((*ScopedWithTransient)(nil)))
		if err != nil {
			t.Fatalf("Failed to resolve scoped service: %v", err)
		}

		scoped2, err := container.Resolve(reflect.TypeOf((*ScopedWithTransient)(nil)))
		if err != nil {
			t.Fatalf("Failed to resolve scoped service: %v", err)
		}

		s1 := scoped1.(*ScopedWithTransient)
		s2 := scoped2.(*ScopedWithTransient)

		// Scoped instances should be different (different resolution chains)
		if s1 == s2 {
			t.Errorf("Expected different scoped instances")
		}

		// But each should have its own transient dependency
		if s1.Transient == s2.Transient {
			t.Errorf("Expected different transient instances in different scoped instances")
		}
	})
}

// TestScopedLifecycleClear tests that scoped instances are properly cleared
func TestScopedLifecycleClear(t *testing.T) {
	container := NewDIContainer()

	// Register a scoped service
	container.RegisterFactory(
		reflect.TypeOf((*ScopedService)(nil)),
		func(c *DIContainer) any {
			return &ScopedService{ID: "scoped"}
		},
		Scoped,
	)

	// Resolve the service to create a scoped instance
	_, err := container.Resolve(reflect.TypeOf((*ScopedService)(nil)))
	if err != nil {
		t.Fatalf("Failed to resolve scoped service: %v", err)
	}

	// Check that scoped instances exist
	container.mu.RLock()
	hasScopedInstances := len(container.scopedInstances) > 0
	container.mu.RUnlock()

	if !hasScopedInstances {
		t.Errorf("Expected scoped instances to exist after resolution")
	}

	// Clear the container
	container.Clear()

	// Check that scoped instances are cleared
	container.mu.RLock()
	hasScopedInstancesAfterClear := len(container.scopedInstances) > 0
	container.mu.RUnlock()

	if hasScopedInstancesAfterClear {
		t.Errorf("Expected scoped instances to be cleared after container clear")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}

// Test service types for testing

type ScopedService struct {
	ID string
}

type SingletonWithScoped struct {
	Scoped *ScopedService
}

type SingletonDependency struct {
	ID string
}

type ScopedWithDependency struct {
	ID        string
	Dep       *SingletonDependency
	CreatedAt int64
}

type ScopedA struct {
	B *ScopedB
}

type ScopedB struct {
	A *ScopedA
}

type SingletonService struct {
	ID string
}

type ScopedWithInjection struct {
	Singleton *SingletonService `inject:""`
}

type TransientService struct {
	ID string
}

type ScopedWithTransient struct {
	Transient *TransientService
}
