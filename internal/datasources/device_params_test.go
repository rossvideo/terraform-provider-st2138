package datasources

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clientpkg "github.com/rossvideo/terraform-provider-st2138/internal/client"
)

func TestDeviceParamsDataSourceMetadataAndSchema(t *testing.T) {
	if NewDeviceParamsDataSource() == nil {
		t.Fatal("NewDeviceParamsDataSource() returned nil")
	}
	dataSource := &deviceParamsDataSource{}
	metadata := &datasource.MetadataResponse{}
	dataSource.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "st2138"}, metadata)
	if metadata.TypeName != "st2138_device_params" {
		t.Fatalf("TypeName = %q, want st2138_device_params", metadata.TypeName)
	}

	response := &datasource.SchemaResponse{}
	dataSource.Schema(context.Background(), datasource.SchemaRequest{}, response)
	if response.Schema.Description == "" {
		t.Fatal("schema description should not be empty")
	}
	if attr, ok := response.Schema.Attributes["slot"].(schema.Int64Attribute); !ok || !attr.Optional {
		t.Error("slot should be an optional int64")
	}
	if attr, ok := response.Schema.Attributes["params"].(schema.MapAttribute); !ok || !attr.Computed {
		t.Error("params should be a computed map")
	}
}

func TestDeviceParamsDataSourceConfigure(t *testing.T) {
	t.Run("nil provider data", func(t *testing.T) {
		dataSource := &deviceParamsDataSource{}
		response := &datasource.ConfigureResponse{}
		dataSource.Configure(context.Background(), datasource.ConfigureRequest{}, response)
		if response.Diagnostics.HasError() || dataSource.client != nil {
			t.Fatalf("Configure() diagnostics = %v, client = %v", response.Diagnostics, dataSource.client)
		}
	})

	t.Run("unexpected provider data", func(t *testing.T) {
		dataSource := &deviceParamsDataSource{}
		response := &datasource.ConfigureResponse{}
		dataSource.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, response)
		if !response.Diagnostics.HasError() {
			t.Fatal("Configure() should reject an unexpected provider data type")
		}
	})

	t.Run("client provider data", func(t *testing.T) {
		dataSource := &deviceParamsDataSource{}
		client := &clientpkg.Client{Endpoint: "localhost:9080", Transport: "rest"}
		response := &datasource.ConfigureResponse{}
		dataSource.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: client}, response)
		if response.Diagnostics.HasError() || dataSource.client != client {
			t.Fatalf("Configure() diagnostics = %v, client = %v", response.Diagnostics, dataSource.client)
		}
	})
}

func TestDeviceParamsDataSourceRead(t *testing.T) {
	ctx := context.Background()
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"slot":1,"params":{"counter":{"type":"INT32","value":{"int32_value":7}}}}`)
	}))
	defer server.Close()

	dataSource := &deviceParamsDataSource{client: &clientpkg.Client{Endpoint: server.URL, Transport: "rest"}}
	schemaResponse := &datasource.SchemaResponse{}
	dataSource.Schema(ctx, datasource.SchemaRequest{}, schemaResponse)
	request := datasource.ReadRequest{Config: newDataSourceConfig(t, schemaResponse.Schema, types.Int64Value(1))}
	response := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	dataSource.Read(ctx, request, response)
	if response.Diagnostics.HasError() {
		t.Fatalf("Read() diagnostics: %v", response.Diagnostics)
	}
	if requestedPath != "/st2138-api/v1/1" {
		t.Errorf("requested path = %q, want /st2138-api/v1/1", requestedPath)
	}
	var state deviceParamsModel
	if diags := response.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("read data-source state: %v", diags)
	}
	if state.ID.ValueString() != server.URL+"/slot/1" {
		t.Errorf("state ID = %q, want endpoint/slot ID", state.ID.ValueString())
	}
	if state.Params.Elements()["counter"].(types.String).ValueString() != "7" {
		t.Errorf("counter param = %v, want 7", state.Params.Elements()["counter"])
	}
}

func TestDeviceParamsDataSourceReadDefaultsSlotAndReportsError(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/st2138-api/v1/0" {
			t.Errorf("requested path = %q, want slot zero", r.URL.Path)
		}
		http.Error(w, "device unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	dataSource := &deviceParamsDataSource{client: &clientpkg.Client{Endpoint: server.URL, Transport: "rest"}}
	schemaResponse := &datasource.SchemaResponse{}
	dataSource.Schema(ctx, datasource.SchemaRequest{}, schemaResponse)
	request := datasource.ReadRequest{Config: newDataSourceConfig(t, schemaResponse.Schema, types.Int64Null())}
	response := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	dataSource.Read(ctx, request, response)
	if !response.Diagnostics.HasError() {
		t.Fatal("Read() should report the REST failure")
	}
}

func newDataSourceConfig(t *testing.T, sourceSchema schema.Schema, slot types.Int64) tfsdk.Config {
	t.Helper()
	ctx := context.Background()
	object := types.ObjectValueMust(map[string]attr.Type{
		"id":     types.StringType,
		"slot":   types.Int64Type,
		"params": types.MapType{ElemType: types.StringType},
	}, map[string]attr.Value{
		"id":     types.StringNull(),
		"slot":   slot,
		"params": types.MapNull(types.StringType),
	})
	raw, err := object.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("convert data-source config: %v", err)
	}
	return tfsdk.Config{Schema: sourceSchema, Raw: raw}
}
