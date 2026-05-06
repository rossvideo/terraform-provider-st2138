package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	clientpkg "github.com/rossvideo/terraform-provider-st2138/internal/client"
)

// NewDeviceParamsDataSource returns a new data source for reading Catena device params.
func NewDeviceParamsDataSource() datasource.DataSource {
	return &deviceParamsDataSource{}
}

type deviceParamsDataSource struct {
	client *clientpkg.Client
}

type deviceParamsModel struct {
	ID     types.String `tfsdk:"id"`
	Slot   types.Int64  `tfsdk:"slot"`
	Params types.Map    `tfsdk:"params"`
}

func (d *deviceParamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_params"
}

func (d *deviceParamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the device model from a Catena device and returns a map of fully-qualified parameter OIDs to their type names (e.g. INT32, STRING, FLOAT32).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier composed of the endpoint and slot.",
			},
			"slot": schema.Int64Attribute{
				Optional:    true,
				Description: "Device slot number to query (default 0).",
			},
			"params": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Map of fully-qualified OID (foid) to ParamType string, e.g. \"brightness\" = \"INT32\".",
			},
		},
	}
}

func (d *deviceParamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cli, ok := req.ProviderData.(*clientpkg.Client)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type",
			fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = cli
}

func (d *deviceParamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceParamsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slot := uint32(0)
	if !state.Slot.IsNull() && !state.Slot.IsUnknown() {
		slot = uint32(state.Slot.ValueInt64())
	}

	rawParams, err := d.client.GetDeviceParams(ctx, slot)
	if err != nil {
		resp.Diagnostics.AddError("failed to read device params",
			fmt.Sprintf("DeviceRequest (slot %d): %s", slot, err))
		return
	}

	// Convert map[string]string → types.Map
	elems := make(map[string]attr.Value, len(rawParams))
	for foid, typeName := range rawParams {
		elems[foid] = types.StringValue(typeName)
	}
	paramsMap, diags := types.MapValue(types.StringType, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/slot/%d", d.client.Endpoint, slot))
	state.Params = paramsMap
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
