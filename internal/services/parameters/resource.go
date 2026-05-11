package parameters

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewParametersResource returns a new parameters resource instance.
func NewParametersResource() resource.Resource {
	return &parametersResource{}
}

type parametersResource struct{}

type parametersModel struct {
	ID             types.String  `tfsdk:"id"`
	Parameters     types.Dynamic `tfsdk:"parameters"`
	ParametersFile types.String  `tfsdk:"parameters_file"`
}

func (r *parametersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_parameters"
}

func (r *parametersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reusable parameter payload that can be referenced by device resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Stable identifier for this parameter payload.",
			},
			"parameters": schema.DynamicAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Parameter payload as an object or list of objects.",
			},
			"parameters_file": schema.StringAttribute{
				Optional:    true,
				Description: "Optional file path to a parameter payload.",
			},
		},
	}
}

func (r *parametersResource) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
}

func normalizeParameters(plan *parametersModel) error {
	hasParameters := !plan.Parameters.IsNull() && !plan.Parameters.IsUnknown() && plan.Parameters.UnderlyingValue() != nil
	hasFile := !plan.ParametersFile.IsNull() && plan.ParametersFile.ValueString() != ""
	if !hasParameters && !hasFile {
		return fmt.Errorf("at least one of parameters or parameters_file must be provided")
	}
	if plan.ID.IsNull() || plan.ID.ValueString() == "" {
		if hasFile {
			plan.ID = types.StringValue("file:" + plan.ParametersFile.ValueString())
		} else {
			plan.ID = types.StringValue("inline")
		}
	}
	return nil
}

func (r *parametersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan parametersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := normalizeParameters(&plan); err != nil {
		resp.Diagnostics.AddError("invalid parameters resource", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *parametersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state parametersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *parametersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan parametersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := normalizeParameters(&plan); err != nil {
		resp.Diagnostics.AddError("invalid parameters resource", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *parametersResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}
