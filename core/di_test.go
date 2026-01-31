package core

import (
	"reflect"
	"sync"
	"testing"
)

type diService interface {
	Name() string
}

type diImpl struct {
	name string
}

func (d *diImpl) Name() string {
	return d.name
}

type diTarget struct {
	Dep diService `inject:""`
}

type diNamedTarget struct {
	Dep diService `inject:"diImpl"`
}

type diNamedPointerTarget struct {
	Dep diService `inject:"*core.diImpl"`
}

type diValueTarget struct {
	Dep diImpl `inject:""`
}

func TestDIInjectInterface(t *testing.T) {
	c := NewDIContainer()
	c.Register(&diImpl{name: "ok"})

	var target diTarget
	if err := c.Inject(&target); err != nil {
		t.Fatalf("inject failed: %v", err)
	}

	if target.Dep == nil || target.Dep.Name() != "ok" {
		t.Fatalf("unexpected injected dependency: %#v", target.Dep)
	}
}

func TestDIInjectByName(t *testing.T) {
	c := NewDIContainer()
	c.Register(&diImpl{name: "named"})

	var target diNamedTarget
	if err := c.Inject(&target); err != nil {
		t.Fatalf("inject failed: %v", err)
	}

	if target.Dep == nil || target.Dep.Name() != "named" {
		t.Fatalf("unexpected injected dependency: %#v", target.Dep)
	}
}

func TestDIInjectInterfaceWithFactory(t *testing.T) {
	c := NewDIContainer()
	c.RegisterFactory(reflect.TypeOf(&diImpl{}), func(c *DIContainer) any {
		return &diImpl{name: "ok"}
	}, Transient)

	var target diTarget
	if err := c.Inject(&target); err != nil {
		t.Fatalf("inject failed: %v", err)
	}

	if target.Dep == nil || target.Dep.Name() != "ok" {
		t.Fatalf("unexpected injected dependency: %#v", target.Dep)
	}
}

func TestDIInjectByNameWithFactory(t *testing.T) {
	c := NewDIContainer()
	c.RegisterFactory(reflect.TypeOf(&diImpl{}), func(c *DIContainer) any {
		return &diImpl{name: "named"}
	}, Singleton)

	var target diNamedPointerTarget
	if err := c.Inject(&target); err != nil {
		t.Fatalf("inject failed: %v", err)
	}

	if target.Dep == nil || target.Dep.Name() != "named" {
		t.Fatalf("unexpected injected dependency: %#v", target.Dep)
	}
}

func TestDIInjectValueFromPointer(t *testing.T) {
	c := NewDIContainer()
	c.Register(&diImpl{name: "value"})

	var target diValueTarget
	if err := c.Inject(&target); err != nil {
		t.Fatalf("inject failed: %v", err)
	}

	if target.Dep.name != "value" {
		t.Fatalf("unexpected injected value: %s", target.Dep.name)
	}
}

func TestDIResolveConcurrentSameType(t *testing.T) {
	c := NewDIContainer()

	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once

	c.RegisterFactory(reflect.TypeOf(&diImpl{}), func(c *DIContainer) any {
		once.Do(func() { close(started) })
		<-release
		return &diImpl{name: "ok"}
	}, Transient)

	results := make(chan error, 2)
	go func() {
		_, err := c.Resolve(reflect.TypeOf(&diImpl{}))
		results <- err
	}()

	<-started
	go func() {
		_, err := c.Resolve(reflect.TypeOf(&diImpl{}))
		results <- err
	}()

	close(release)

	for i := 0; i < 2; i++ {
		if err := <-results; err != nil {
			t.Fatalf("concurrent resolve failed: %v", err)
		}
	}
}
