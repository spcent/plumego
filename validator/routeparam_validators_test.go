package validator

import "testing"

func TestRouteParamRangeValidatorStrictParse(t *testing.T) {
	v := RouteParamNewRange(1, 100)
	if err := v.Validate("id", "12abc"); err == nil {
		t.Fatal("expected strict integer parsing error")
	}
}

func TestRouteParamRangeValidatorInvalidRange(t *testing.T) {
	v := RouteParamNewRange(10, 1)
	if err := v.Validate("id", "5"); err == nil {
		t.Fatal("expected invalid range configuration error")
	}
}

func TestRouteParamCustomValidatorNilFunction(t *testing.T) {
	v := RouteParamNewCustom(nil)
	if err := v.Validate("name", "alice"); err == nil {
		t.Fatal("expected nil custom validator error")
	}
}
