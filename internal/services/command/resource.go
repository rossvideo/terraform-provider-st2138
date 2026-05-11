package command

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewCommandResource returns a new command resource instance.
func NewCommandResource() resource.Resource {
	return &commandResource{}
}

type commandResource struct{}

type commandModel struct {
	ID                      types.String  `tfsdk:"id"`
	Command                 types.String  `tfsdk:"command"`
	Value                   types.Dynamic `tfsdk:"value"`
	StatusFoid              types.String  `tfsdk:"status_foid"`
	StatusSuccessValue      types.String  `tfsdk:"status_success_value"`
	StatusSuccessComparator types.String  `tfsdk:"status_success_comparator"`
	TimeoutSeconds          types.Int64   `tfsdk:"timeout_seconds"`
}

func (r *commandResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_command"
}

func (r *commandResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reusable command definition that can be referenced by st2138_device startup/shutdown command blocks.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Stable identifier for this command definition.",
			},
			"command": schema.StringAttribute{
				Required:    true,
				Description: "Command OID to execute.",
			},
			"value": schema.DynamicAttribute{
				Optional:    true,
				Description: "Optional command value payload.",
			},
			"status_foid": schema.StringAttribute{
				Optional:    true,
				Description: "Optional parameter OID to poll for success.",
			},
			"status_success_value": schema.StringAttribute{
				Optional:    true,
				Description: "Expected value used by status_success_comparator.",
			},
			"status_success_comparator": schema.StringAttribute{
				Optional:    true,
				Description: "Comparator: eq, ne, gt, lt, ge, le.",
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional:    true,
				Description: "Timeout in seconds. Defaults to 5.",
			},
		},
	}
}

func (r *commandResource) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
}

func normalizeCommand(plan *commandModel) (string, error) {
	command := ""
	if !plan.Command.IsNull() && plan.Command.ValueString() != "" {
		command = plan.Command.ValueString()
	}
	if command == "" {
		return "", fmt.Errorf("command must be provided")
	}
	plan.Command = types.StringValue(command)
	if plan.TimeoutSeconds.IsNull() || plan.TimeoutSeconds.ValueInt64() <= 0 {
		plan.TimeoutSeconds = types.Int64Value(5)
	}
	plan.ID = types.StringValue(command)
	return command, nil
}

func (r *commandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan commandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := normalizeCommand(&plan); err != nil {
		resp.Diagnostics.AddError("invalid command resource", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *commandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state commandModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *commandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan commandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := normalizeCommand(&plan); err != nil {
		resp.Diagnostics.AddError("invalid command resource", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *commandResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {}
