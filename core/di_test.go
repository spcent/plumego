package core

import "testing"

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
