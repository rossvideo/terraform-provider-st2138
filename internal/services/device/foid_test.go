package device

import (
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToFOID(t *testing.T) {
	tests := map[string]string{
		"authz_monitor":       "/authz_monitor",
		"/authz_monitor":      "/authz_monitor",
		"  counter ":          "/counter",
		"constraint/examples": "/constraint/examples",
		"":                    "",
	}
	for in, want := range tests {
		if got := toFOID(in); got != want {
			t.Errorf("toFOID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestChangedParams(t *testing.T) {
	prev := map[string]attr.Value{
		"counter":        types.NumberValue(big.NewFloat(1)),
		"string_example": types.StringValue("Hello"),
		"removed":        types.StringValue("x"),
	}
	next := map[string]attr.Value{
		"counter":        types.NumberValue(big.NewFloat(1)),
		"string_example": types.StringValue("World"),
		"added":          types.StringValue("y"),
	}

	got := changedParams(prev, next)

	if len(got) != 2 {
		t.Fatalf("changed = %v, want string_example and added", got)
	}
	if _, ok := got["string_example"]; !ok {
		t.Error("string_example should be changed")
	}
	if _, ok := got["added"]; !ok {
		t.Error("added should be changed")
	}
}

func TestCommandInvocationFromAttr_NormalizesOIDs(t *testing.T) {
	r := &deviceResource{}

	inv, err := r.commandInvocationFromAttr(types.StringValue("fib_stop"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.OID != "/fib_stop" {
		t.Errorf("OID = %q, want /fib_stop", inv.OID)
	}

	obj := types.ObjectValueMust(
		map[string]attr.Type{"command": types.StringType, "status_foid": types.StringType},
		map[string]attr.Value{"command": types.StringValue("fib_start"), "status_foid": types.StringValue("number_example")},
	)
	inv, err = r.commandInvocationFromAttr(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.OID != "/fib_start" {
		t.Errorf("OID = %q, want /fib_start", inv.OID)
	}
	if inv.StatusFoid != "/number_example" {
		t.Errorf("StatusFoid = %q, want /number_example", inv.StatusFoid)
	}
}
