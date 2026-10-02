package parameters

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeParameters(t *testing.T) {
	inline := types.DynamicValue(types.ObjectValueMust(
		map[string]attr.Type{"counter": types.NumberType},
		map[string]attr.Value{"counter": types.NumberValue(big.NewFloat(1))},
	))

	tests := []struct {
		name    string
		plan    parametersModel
		wantID  string
		wantErr bool
	}{
		{name: "requires payload", plan: parametersModel{Parameters: types.DynamicNull(), ParametersFile: types.StringNull()}, wantErr: true},
		{name: "inline payload id", plan: parametersModel{Parameters: inline, ParametersFile: types.StringNull()}, wantID: "inline"},
		{name: "file payload id", plan: parametersModel{Parameters: types.DynamicNull(), ParametersFile: types.StringValue("params.json")}, wantID: "file:params.json"},
		{name: "explicit id is retained", plan: parametersModel{ID: types.StringValue("custom"), Parameters: inline}, wantID: "custom"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := normalizeParameters(&test.plan)
			if test.wantErr {
				if err == nil {
					t.Fatal("normalizeParameters() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeParameters() error = %v", err)
			}
			if test.plan.ID.ValueString() != test.wantID {
				t.Errorf("ID = %q, want %q", test.plan.ID.ValueString(), test.wantID)
			}
		})
	}
}

func TestParametersResourceMetadataAndNoOpMethods(t *testing.T) {
	parameters := NewParametersResource().(*parametersResource)
	metadata := &resource.MetadataResponse{}
	parameters.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "st2138"}, metadata)
	if metadata.TypeName != "st2138_parameters" {
		t.Fatalf("TypeName = %q, want st2138_parameters", metadata.TypeName)
	}
	parameters.Configure(context.Background(), resource.ConfigureRequest{}, &resource.ConfigureResponse{})
	parameters.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
}

func TestParametersResourceLifecycle(t *testing.T) {
	ctx := context.Background()
	parameters := &parametersResource{}
	schemaResponse := &resource.SchemaResponse{}
	parameters.Schema(ctx, resource.SchemaRequest{}, schemaResponse)
	inlineParameters := types.DynamicValue(types.ObjectValueMust(
		map[string]attr.Type{"counter": types.NumberType},
		map[string]attr.Value{"counter": types.NumberValue(big.NewFloat(4))},
	))
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	planDiagnostics := plan.Set(ctx, &parametersModel{
		Parameters:     inlineParameters,
		ParametersFile: types.StringNull(),
	})
	if planDiagnostics.HasError() {
		t.Fatalf("build parameters plan: %v", planDiagnostics)
	}

	createResponse := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	parameters.Create(ctx, resource.CreateRequest{Plan: plan}, createResponse)
	if createResponse.Diagnostics.HasError() {
		t.Fatalf("Create() diagnostics: %v", createResponse.Diagnostics)
	}
	var created parametersModel
	if diags := createResponse.State.Get(ctx, &created); diags.HasError() {
		t.Fatalf("read created state: %v", diags)
	}
	if created.ID.ValueString() != "inline" {
		t.Errorf("created ID = %q, want inline", created.ID.ValueString())
	}

	readResponse := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	parameters.Read(ctx, resource.ReadRequest{State: createResponse.State}, readResponse)
	if readResponse.Diagnostics.HasError() {
		t.Fatalf("Read() diagnostics: %v", readResponse.Diagnostics)
	}
	updateResponse := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	parameters.Update(ctx, resource.UpdateRequest{Plan: plan}, updateResponse)
	if updateResponse.Diagnostics.HasError() {
		t.Fatalf("Update() diagnostics: %v", updateResponse.Diagnostics)
	}
}

func TestParametersResourceRejectsMissingPayload(t *testing.T) {
	ctx := context.Background()
	parameters := &parametersResource{}
	schemaResponse := &resource.SchemaResponse{}
	parameters.Schema(ctx, resource.SchemaRequest{}, schemaResponse)
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := plan.Set(ctx, &parametersModel{
		Parameters:     types.DynamicNull(),
		ParametersFile: types.StringNull(),
	}); diags.HasError() {
		t.Fatalf("build invalid parameters plan: %v", diags)
	}
	createResponse := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	parameters.Create(ctx, resource.CreateRequest{Plan: plan}, createResponse)
	if !createResponse.Diagnostics.HasError() {
		t.Fatal("Create() should reject missing parameters")
	}
	updateResponse := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	parameters.Update(ctx, resource.UpdateRequest{Plan: plan}, updateResponse)
	if !updateResponse.Diagnostics.HasError() {
		t.Fatal("Update() should reject missing parameters")
	}
}
