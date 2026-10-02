package command

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCommandResourceMetadataAndSchema(t *testing.T) {
	if NewCommandResource() == nil {
		t.Fatal("NewCommandResource() returned nil")
	}
	command := &commandResource{}
	metadata := &resource.MetadataResponse{}
	command.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "st2138"}, metadata)
	if metadata.TypeName != "st2138_command" {
		t.Fatalf("TypeName = %q, want st2138_command", metadata.TypeName)
	}

	response := &resource.SchemaResponse{}
	command.Schema(context.Background(), resource.SchemaRequest{}, response)
	if response.Schema.Description == "" {
		t.Fatal("schema description should not be empty")
	}
	if attr, ok := response.Schema.Attributes["command"].(schema.StringAttribute); !ok || !attr.Required {
		t.Error("command should be a required string")
	}
	if attr, ok := response.Schema.Attributes["id"].(schema.StringAttribute); !ok || !attr.Computed {
		t.Error("id should be a computed string")
	}
	if attr, ok := response.Schema.Attributes["value"].(schema.DynamicAttribute); !ok || !attr.Optional {
		t.Error("value should be an optional dynamic attribute")
	}
}

func TestCommandResourceNoOpMethods(t *testing.T) {
	command := &commandResource{}
	command.Configure(context.Background(), resource.ConfigureRequest{}, &resource.ConfigureResponse{})
	command.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
}

func TestNormalizeCommand(t *testing.T) {
	tests := []struct {
		name        string
		plan        commandModel
		wantCommand string
		wantID      string
		wantTimeout int64
		wantError   bool
	}{
		{
			name:        "defaults timeout and id",
			plan:        commandModel{Command: types.StringValue("reset"), TimeoutSeconds: types.Int64Null()},
			wantCommand: "reset",
			wantID:      "reset",
			wantTimeout: 5,
		},
		{
			name:        "keeps explicit id and timeout",
			plan:        commandModel{ID: types.StringValue("stable-id"), Command: types.StringValue("start"), TimeoutSeconds: types.Int64Value(9)},
			wantCommand: "start",
			wantID:      "start",
			wantTimeout: 9,
		},
		{
			name:        "defaults nonpositive timeout",
			plan:        commandModel{Command: types.StringValue("stop"), TimeoutSeconds: types.Int64Value(0)},
			wantCommand: "stop",
			wantID:      "stop",
			wantTimeout: 5,
		},
		{
			name:      "missing command",
			plan:      commandModel{Command: types.StringNull(), TimeoutSeconds: types.Int64Null()},
			wantError: true,
		},
		{
			name:      "empty command",
			plan:      commandModel{Command: types.StringValue(""), TimeoutSeconds: types.Int64Null()},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandValue, err := normalizeCommand(&test.plan)
			if test.wantError {
				if err == nil {
					t.Fatal("normalizeCommand() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeCommand() error = %v", err)
			}
			if commandValue != test.wantCommand || test.plan.ID.ValueString() != test.wantID {
				t.Errorf("command/id = %q/%q, want %q/%q", commandValue, test.plan.ID.ValueString(), test.wantCommand, test.wantID)
			}
			if test.plan.TimeoutSeconds.ValueInt64() != test.wantTimeout {
				t.Errorf("timeout = %d, want %d", test.plan.TimeoutSeconds.ValueInt64(), test.wantTimeout)
			}
		})
	}
}

func TestCommandResourceLifecycle(t *testing.T) {
	ctx := context.Background()
	command := &commandResource{}
	schemaResponse := &resource.SchemaResponse{}
	command.Schema(ctx, resource.SchemaRequest{}, schemaResponse)
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	planDiagnostics := plan.Set(ctx, &commandModel{
		Command:                 types.StringValue("reset"),
		Value:                   types.DynamicNull(),
		StatusFoid:              types.StringNull(),
		StatusSuccessValue:      types.StringNull(),
		StatusSuccessComparator: types.StringNull(),
		TimeoutSeconds:          types.Int64Null(),
	})
	if planDiagnostics.HasError() {
		t.Fatalf("build command plan: %v", planDiagnostics)
	}

	createResponse := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	command.Create(ctx, resource.CreateRequest{Plan: plan}, createResponse)
	if createResponse.Diagnostics.HasError() {
		t.Fatalf("Create() diagnostics: %v", createResponse.Diagnostics)
	}
	var created commandModel
	if diags := createResponse.State.Get(ctx, &created); diags.HasError() {
		t.Fatalf("read created state: %v", diags)
	}
	if created.ID.ValueString() != "reset" || created.TimeoutSeconds.ValueInt64() != 5 {
		t.Errorf("created command state = %#v", created)
	}

	readResponse := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	command.Read(ctx, resource.ReadRequest{State: createResponse.State}, readResponse)
	if readResponse.Diagnostics.HasError() {
		t.Fatalf("Read() diagnostics: %v", readResponse.Diagnostics)
	}
	updateResponse := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	command.Update(ctx, resource.UpdateRequest{Plan: plan}, updateResponse)
	if updateResponse.Diagnostics.HasError() {
		t.Fatalf("Update() diagnostics: %v", updateResponse.Diagnostics)
	}
}

func TestCommandResourceRejectsEmptyCommand(t *testing.T) {
	ctx := context.Background()
	command := &commandResource{}
	schemaResponse := &resource.SchemaResponse{}
	command.Schema(ctx, resource.SchemaRequest{}, schemaResponse)
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := plan.Set(ctx, &commandModel{
		Command:                 types.StringValue(""),
		Value:                   types.DynamicNull(),
		StatusFoid:              types.StringNull(),
		StatusSuccessValue:      types.StringNull(),
		StatusSuccessComparator: types.StringNull(),
		TimeoutSeconds:          types.Int64Null(),
	}); diags.HasError() {
		t.Fatalf("build invalid command plan: %v", diags)
	}
	createResponse := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	command.Create(ctx, resource.CreateRequest{Plan: plan}, createResponse)
	if !createResponse.Diagnostics.HasError() {
		t.Fatal("Create() should reject an empty command")
	}
	updateResponse := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	command.Update(ctx, resource.UpdateRequest{Plan: plan}, updateResponse)
	if !updateResponse.Diagnostics.HasError() {
		t.Fatal("Update() should reject an empty command")
	}
}
