package device

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestDeviceResourceRESTLifecycle(t *testing.T) {
	ctx := context.Background()
	var counter int32 = 1
	var setCount, commandCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/st2138-api/v1/0":
			device := &st2138pb.Device{
				Slot: 0,
				Params: map[string]*st2138pb.Param{
					"counter": {Type: st2138pb.ParamType_INT32, Value: &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: counter}}},
				},
				Commands: map[string]*st2138pb.Param{
					"reset": {Value: &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}}},
				},
			}
			writeDeviceTestProto(t, w, http.StatusOK, device)
		case req.Method == http.MethodGet && req.URL.Path == "/st2138-api/v1/0/param/counter":
			writeDeviceTestProto(t, w, http.StatusOK, &st2138pb.DeviceComponent_ComponentParam{
				Oid:   "counter",
				Param: &st2138pb.Param{Type: st2138pb.ParamType_INT32},
			})
		case req.Method == http.MethodGet && req.URL.Path == "/st2138-api/v1/0/value/counter":
			writeDeviceTestProto(t, w, http.StatusOK, &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: counter}})
		case req.Method == http.MethodPut && req.URL.Path == "/st2138-api/v1/0/value/counter":
			value := &st2138pb.Value{}
			if err := (protojson.UnmarshalOptions{}).Unmarshal(readDeviceTestBody(t, req), value); err != nil {
				t.Errorf("decode SetValue body: %v", err)
			} else {
				counter = value.GetInt32Value()
				setCount++
			}
			w.WriteHeader(http.StatusNoContent)
		case req.Method == http.MethodPost && (req.URL.Path == "/st2138-api/v1/0/command/reset" || req.URL.Path == "/st2138-api/v1/0/command/stop"):
			if req.URL.Query().Get("respond") != "true" {
				t.Errorf("command respond query = %q, want true", req.URL.Query().Get("respond"))
			}
			commandCount++
			writeDeviceTestProto(t, w, http.StatusOK, &st2138pb.CommandResponse{
				Kind: &st2138pb.CommandResponse_NoResponse{NoResponse: &st2138pb.Empty{}},
			})
		default:
			http.Error(w, fmt.Sprintf("unexpected request %s %s", req.Method, req.URL.Path), http.StatusNotFound)
		}
	}))
	defer server.Close()

	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	port, err := strconv.Atoi(endpoint.Port())
	if err != nil {
		t.Fatalf("parse test server port: %v", err)
	}

	deviceResource := &deviceResource{}
	schemaResponse := &resource.SchemaResponse{}
	deviceResource.Schema(ctx, resource.SchemaRequest{}, schemaResponse)
	parameters := deviceTestParameters(2)
	startupCommands := deviceTestStartupCommands()
	plan := deviceTestPlan(t, schemaResponse.Schema, endpoint.Hostname(), int64(port), parameters, startupCommands)

	createResponse := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	deviceResource.Create(ctx, resource.CreateRequest{Plan: plan}, createResponse)
	if createResponse.Diagnostics.HasError() {
		t.Fatalf("Create() diagnostics: %v", createResponse.Diagnostics)
	}
	if counter != 2 || commandCount != 1 || setCount != 1 {
		t.Fatalf("after create: counter=%d, sets=%d, commands=%d", counter, setCount, commandCount)
	}
	var created deviceModel
	if diags := createResponse.State.Get(ctx, &created); diags.HasError() {
		t.Fatalf("read created state: %v", diags)
	}
	if created.StatusValue.ValueString() != "2" {
		t.Errorf("create status_value = %q, want 2", created.StatusValue.ValueString())
	}

	readResponse := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	deviceResource.Read(ctx, resource.ReadRequest{State: createResponse.State}, readResponse)
	if readResponse.Diagnostics.HasError() {
		t.Fatalf("Read() diagnostics: %v", readResponse.Diagnostics)
	}

	updatePlan := deviceTestPlan(t, schemaResponse.Schema, endpoint.Hostname(), int64(port), deviceTestParameters(3), startupCommands)
	updateResponse := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	deviceResource.Update(ctx, resource.UpdateRequest{Plan: updatePlan, State: createResponse.State}, updateResponse)
	if updateResponse.Diagnostics.HasError() {
		t.Fatalf("Update() diagnostics: %v", updateResponse.Diagnostics)
	}
	if counter != 3 || setCount != 2 {
		t.Errorf("after update: counter=%d, sets=%d, want counter=3 and 2 sets", counter, setCount)
	}

	deleteResponse := &resource.DeleteResponse{}
	deviceResource.Delete(ctx, resource.DeleteRequest{State: updateResponse.State}, deleteResponse)
	if deleteResponse.Diagnostics.HasError() {
		t.Fatalf("Delete() diagnostics: %v", deleteResponse.Diagnostics)
	}
	if commandCount != 2 {
		t.Errorf("command count after delete = %d, want startup and shutdown commands", commandCount)
	}
}

func deviceTestParameters(counter int64) types.Dynamic {
	object := types.ObjectValueMust(
		map[string]attr.Type{"counter": types.NumberType},
		map[string]attr.Value{"counter": types.NumberValue(big.NewFloat(float64(counter)))},
	)
	return types.DynamicValue(object)
}

func deviceTestStartupCommands() *commandRefsBlockModel {
	commandTypes := map[string]attr.Type{
		"command":                   types.StringType,
		"value":                     types.DynamicType,
		"status_foid":               types.StringType,
		"status_success_value":      types.StringType,
		"status_success_comparator": types.StringType,
		"timeout_seconds":           types.NumberType,
	}
	commandObject := types.ObjectType{AttrTypes: commandTypes}
	command := types.ObjectValueMust(commandTypes, map[string]attr.Value{
		"command":                   types.StringValue("reset"),
		"value":                     types.DynamicNull(),
		"status_foid":               types.StringValue("counter"),
		"status_success_value":      types.StringValue("2"),
		"status_success_comparator": types.StringValue("eq"),
		"timeout_seconds":           types.NumberValue(big.NewFloat(1)),
	})
	commandList := types.TupleValueMust([]attr.Type{commandObject}, []attr.Value{command})
	return &commandRefsBlockModel{Commands: types.DynamicValue(commandList)}
}

func deviceTestPlan(t *testing.T, deviceSchema schema.Schema, address string, port int64, parameters types.Dynamic, startupCommands *commandRefsBlockModel) tfsdk.Plan {
	t.Helper()
	network := types.ObjectValueMust(networkAttrTypes, map[string]attr.Value{
		"address":   types.StringValue(address),
		"port":      types.Int64Value(port),
		"transport": types.StringValue("rest"),
		"tls":       types.BoolValue(false),
	})
	model := &deviceModel{
		ID:                types.StringUnknown(),
		Name:              types.StringValue("coverage-test"),
		SlotID:            types.Int64Value(0),
		Network:           network,
		Parameters:        parameters,
		ParametersOut:     types.MapUnknown(types.StringType),
		FullParametersOut: types.MapUnknown(types.StringType),
		CommandsOut:       types.MapUnknown(types.StringType),
		StatusValue:       types.StringUnknown(),
		StartupCommands:   startupCommands,
		ShutdownCommands:  deviceTestShutdownCommands(),
	}
	plan := tfsdk.Plan{Schema: deviceSchema}
	if diags := plan.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("build device plan: %v", diags)
	}
	return plan
}

func deviceTestShutdownCommands() *commandRefsBlockModel {
	commands := types.TupleValueMust([]attr.Type{types.StringType}, []attr.Value{types.StringValue("stop")})
	return &commandRefsBlockModel{Commands: types.DynamicValue(commands)}
}

func writeDeviceTestProto(t *testing.T, writer http.ResponseWriter, status int, message proto.Message) {
	t.Helper()
	body, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(message)
	if err != nil {
		t.Fatalf("marshal test response: %v", err)
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_, _ = writer.Write(body)
}

func readDeviceTestBody(t *testing.T, request *http.Request) []byte {
	t.Helper()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	return body
}

func TestDeviceModel_Initialization(t *testing.T) {
	model := &deviceModel{
		Name: types.StringValue("test-device"),
	}

	if model.Name.ValueString() != "test-device" {
		t.Errorf("Name = %s, want test-device", model.Name.ValueString())
	}
}

func TestDeviceModel_NullValues(t *testing.T) {
	model := &deviceModel{
		ID:   types.StringNull(),
		Name: types.StringNull(),
	}

	if !model.ID.IsNull() {
		t.Error("ID should be null")
	}
	if !model.Name.IsNull() {
		t.Error("Name should be null")
	}
}

func TestDeviceModel_UnknownValues(t *testing.T) {
	model := &deviceModel{
		ID:   types.StringUnknown(),
		Name: types.StringUnknown(),
	}

	if !model.ID.IsUnknown() {
		t.Error("ID should be unknown")
	}
	if !model.Name.IsUnknown() {
		t.Error("Name should be unknown")
	}
}

func TestParamPairModel(t *testing.T) {
	pair := paramPairModel{
		Oid:   types.StringValue("test/oid"),
		Value: types.StringValue("test-value"),
	}

	if pair.Oid.ValueString() != "test/oid" {
		t.Errorf("Oid = %s, want /test/oid", pair.Oid.ValueString())
	}
	if pair.Value.ValueString() != "test-value" {
		t.Errorf("Value = %s, want test-value", pair.Value.ValueString())
	}
}

func TestDeviceStatusModel(t *testing.T) {
	status := &deviceStatusModel{
		Oid:        types.StringValue("status/ready"),
		ReadyValue: types.StringValue("true"),
	}

	if status.Oid.ValueString() != "status/ready" {
		t.Errorf("Oid = %s, want /status/ready", status.Oid.ValueString())
	}
	if status.ReadyValue.ValueString() != "true" {
		t.Errorf("ReadyValue = %s, want true", status.ReadyValue.ValueString())
	}
}

func TestDeviceResource_PathExists(t *testing.T) {
	r := &deviceResource{}

	// Test with current directory (should exist)
	if !r.pathExists(".") {
		t.Error("Current directory should exist")
	}

	// Test with non-existent path
	if r.pathExists("/nonexistent/path/12345") {
		t.Error("Non-existent path should return false")
	}
}

func TestDeviceResource_SelectHostPortForInternal(t *testing.T) {
	r := &deviceResource{}

	tests := []struct {
		name     string
		ports    []string
		internal int
		want     int
	}{
		{
			name:     "simple mapping",
			ports:    []string{"7254:6254"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "with ip address",
			ports:    []string{"127.0.0.1:7254:6254"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "with protocol",
			ports:    []string{"7254:6254/tcp"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "no match",
			ports:    []string{"8000:8080"},
			internal: 6254,
			want:     0,
		},
		{
			name:     "empty ports",
			ports:    []string{},
			internal: 6254,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.selectHostPortForInternal(tt.ports, tt.internal)
			if got != tt.want {
				t.Errorf("selectHostPortForInternal() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDeviceResource_ParseValueString(t *testing.T) {
	r := &deviceResource{}

	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{
			name:     "boolean true",
			input:    "true",
			wantType: "bool",
		},
		{
			name:     "boolean false",
			input:    "false",
			wantType: "bool",
		},
		{
			name:     "integer",
			input:    "42",
			wantType: "float64",
		},
		{
			name:     "float",
			input:    "3.14",
			wantType: "float64",
		},
		{
			name:     "string",
			input:    "hello",
			wantType: "string",
		},
		{
			name:     "empty",
			input:    "",
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.parseValueString(tt.input)
			gotType := ""
			switch result.(type) {
			case bool:
				gotType = "bool"
			case float64:
				gotType = "float64"
			case string:
				gotType = "string"
			}
			if gotType != tt.wantType {
				t.Errorf("parseValueString(%q) type = %s, want %s", tt.input, gotType, tt.wantType)
			}
		})
	}
}

func TestDeviceResource_GetContainerID(t *testing.T) {
	r := &deviceResource{}

	// Test with empty name
	id := r.getContainerID("")
	if id != "" {
		t.Errorf("getContainerID(\"\") = %s, want empty string", id)
	}

	// Test with whitespace name
	id = r.getContainerID("   ")
	if id != "" {
		t.Errorf("getContainerID(\"   \") = %s, want empty string", id)
	}
}
