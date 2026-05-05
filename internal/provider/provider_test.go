package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCatenaProviderModel(t *testing.T) {
	model := catenaProviderModel{
		Endpoint:       types.StringValue("localhost:6254"),
		Transport:      types.StringValue("grpc"),
		DevicesDir:     types.StringValue("/devices"),
		ExecutablesDir: types.StringValue("/executables"),
	}

	if model.Endpoint.ValueString() != "localhost:6254" {
		t.Errorf("Endpoint = %s, want localhost:6254", model.Endpoint.ValueString())
	}
	if model.Transport.ValueString() != "grpc" {
		t.Errorf("Transport = %s, want grpc", model.Transport.ValueString())
	}
}

func TestCatenaProviderModel_NullValues(t *testing.T) {
	model := catenaProviderModel{
		Endpoint:  types.StringNull(),
		Transport: types.StringNull(),
	}

	if !model.Endpoint.IsNull() {
		t.Error("Endpoint should be null")
	}
	if !model.Transport.IsNull() {
		t.Error("Transport should be null")
	}
}

func TestNew(t *testing.T) {
	provider := New()
	if provider == nil {
		t.Error("New() should return a provider instance")
	}

	// Type assertion to verify it's the correct type
	_, ok := provider.(*catenaProvider)
	if !ok {
		t.Error("New() should return a *catenaProvider")
	}
}
