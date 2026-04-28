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

func TestRouteParamCompositeValidatorNilChild(t *testing.T) {
	v := RouteParamNewComposite(RouteParamNotEmpty, nil)
	if err := v.Validate("name", "alice"); err == nil {
		t.Fatal("expected nil child validator error")
	}
}

func TestRouteParamLengthValidatorCountsRunes(t *testing.T) {
	v := RouteParamNewLength(2, 2)
	if err := v.Validate("name", "世界"); err != nil {
		t.Fatalf("expected two unicode characters to be valid length, got %v", err)
	}
}
