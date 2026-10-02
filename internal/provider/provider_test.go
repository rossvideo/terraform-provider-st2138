package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clientpkg "github.com/rossvideo/terraform-provider-st2138/internal/client"
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

func TestProviderMetadataSchemaAndRegistrations(t *testing.T) {
	p := &catenaProvider{}
	metadata := &frameworkprovider.MetadataResponse{}
	p.Metadata(context.Background(), frameworkprovider.MetadataRequest{}, metadata)
	if metadata.TypeName != "st2138" {
		t.Fatalf("TypeName = %q, want st2138", metadata.TypeName)
	}

	schemaResponse := &frameworkprovider.SchemaResponse{}
	p.Schema(context.Background(), frameworkprovider.SchemaRequest{}, schemaResponse)
	if schemaResponse.Schema.Description == "" {
		t.Fatal("provider schema description should not be empty")
	}
	for _, name := range []string{"endpoint", "transport", "devices_dir", "executables_dir"} {
		if _, ok := schemaResponse.Schema.Attributes[name]; !ok {
			t.Errorf("provider schema is missing %q", name)
		}
	}

	resources := p.Resources(context.Background())
	if len(resources) != 3 {
		t.Fatalf("resource count = %d, want 3", len(resources))
	}
	for index, factory := range resources {
		if factory() == nil {
			t.Errorf("resource factory %d returned nil", index)
		}
	}
	datasources := p.DataSources(context.Background())
	if len(datasources) != 1 || datasources[0]() == nil {
		t.Fatalf("data source factories = %d, want one non-nil factory", len(datasources))
	}
}

func TestProviderConfigure(t *testing.T) {
	ctx := context.Background()
	p := &catenaProvider{}
	schemaResponse := &frameworkprovider.SchemaResponse{}
	p.Schema(ctx, frameworkprovider.SchemaRequest{}, schemaResponse)

	rawObject := types.ObjectValueMust(map[string]attr.Type{
		"endpoint":        types.StringType,
		"transport":       types.StringType,
		"devices_dir":     types.StringType,
		"executables_dir": types.StringType,
	}, map[string]attr.Value{
		"endpoint":        types.StringValue("localhost:9080"),
		"transport":       types.StringValue("rest"),
		"devices_dir":     types.StringValue("/devices"),
		"executables_dir": types.StringValue("/executables"),
	})
	rawConfig, err := rawObject.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("convert provider config: %v", err)
	}
	response := &frameworkprovider.ConfigureResponse{}
	p.Configure(ctx, frameworkprovider.ConfigureRequest{
		Config: tfsdk.Config{Schema: schemaResponse.Schema, Raw: rawConfig},
	}, response)
	if response.Diagnostics.HasError() {
		t.Fatalf("Configure() diagnostics: %v", response.Diagnostics)
	}
	configuredClient, ok := response.ResourceData.(*clientpkg.Client)
	if !ok {
		t.Fatalf("ResourceData type = %T, want *client.Client", response.ResourceData)
	}
	if configuredClient.Endpoint != "localhost:9080" || configuredClient.Transport != "rest" || configuredClient.DevicesDir != "/executables" {
		t.Errorf("configured client = %#v", configuredClient)
	}
}
