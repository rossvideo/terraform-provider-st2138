package device

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	clientpkg "github.com/rossvideo/terraform-provider-st2138/internal/client"
	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// NewDeviceResource returns a new device resource instance.
func NewDeviceResource() resource.Resource {
	return &deviceResource{}
}

type deviceResource struct {
	client *clientpkg.Client
}

// commandRefsBlockModel stores reusable command refs for startup/shutdown execution.
type commandRefsBlockModel struct {
	Commands types.Dynamic `tfsdk:"commands"`
}

// deviceModel holds the Terraform state for a st2138_device resource.
type deviceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	SlotID  types.Int64  `tfsdk:"slot"`
	Network types.Object `tfsdk:"network"`
	// parameters: dynamic value matching a single-slot shape, e.g. [{"counter": 1}]
	Parameters        types.Dynamic          `tfsdk:"parameters"`
	ParametersOut     types.Map              `tfsdk:"parameters_out"`
	FullParametersOut types.Map              `tfsdk:"full_parameters_out"`
	CommandsOut       types.Map              `tfsdk:"commands_out"`
	StatusValue       types.String           `tfsdk:"status_value"`
	StartupCommands   *commandRefsBlockModel `tfsdk:"startup_commands"`
	ShutdownCommands  *commandRefsBlockModel `tfsdk:"shutdown_commands"`

	// Legacy fields - commented out, kept for test compatibility via separate structs below.
	// DeviceType   types.String       `tfsdk:"device_type"`
	// ContainerID  types.String       `tfsdk:"container_id"`
	// Params       []paramPairModel   `tfsdk:"params"`
	// ParamsMap    types.Map          `tfsdk:"params_map"`
	// SlotID       types.Int64        `tfsdk:"slot_id"`
	// StartCommand types.String       `tfsdk:"start_command"`
	// StopCommand  types.String       `tfsdk:"stop_command"`
	// DeviceStatus *deviceStatusModel `tfsdk:"device_status"`
	// StatusValue  types.String       `tfsdk:"status_value"`
	// ApplyAll     types.Bool         `tfsdk:"apply_all"`
	// Address      types.String       `tfsdk:"address"`
	// Port         types.Int64        `tfsdk:"port"`
}

var networkAttrTypes = map[string]attr.Type{
	"address":   types.StringType,
	"port":      types.Int64Type,
	"transport": types.StringType,
	"tls":       types.BoolType,
}

// deviceStatusModel kept for backward compatibility with existing tests.
type deviceStatusModel struct {
	Endpoint   types.String `tfsdk:"endpoint"`
	Oid        types.String `tfsdk:"oid"`
	ReadyValue types.String `tfsdk:"ready_value"`
}

// paramPairModel kept for backward compatibility with existing tests.
type paramPairModel struct {
	Oid   types.String `tfsdk:"oid"`
	Value types.String `tfsdk:"value"`
}

// Create: Handles device creation in Terraform
// Read: Fetches current device state from backend
// Update: Modifies existing device on backend
// Delete: Removes device from backend
func (r *deviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *deviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Catena device: sets parameters via gRPC on create and reads the device model.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique identifier for the device.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Human-readable device name.",
			},
			"slot": schema.Int64Attribute{
				Required:    true,
				Description: "Device slot this resource manages.",
			},
			"parameters": schema.DynamicAttribute{
				Optional:    true,
				Description: "Dynamic parameter set for this slot. All are applied on create; only changed values are applied on update. Supports object or list-of-objects shapes such as [{\"counter\": 1, \"struct_example\": {...}}].",
			},
			"parameters_out": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Writable parameters for this slot from DeviceRequest.",
			},
			"full_parameters_out": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "All parameters for this slot (including read-only) from DeviceRequest.",
			},
			"commands_out": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Commands for this slot from DeviceRequest.",
			},
			"status_value": schema.StringAttribute{
				Computed:      true,
				Description:   "Most recent status value observed from command status polling when available.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
		Blocks: map[string]schema.Block{
			"network": schema.SingleNestedBlock{
				Description: "Network configuration for this slot.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Required:    true,
						Description: "Host/IP for the Catena endpoint.",
					},
					"port": schema.Int64Attribute{
						Required:    true,
						Description: "Port for the Catena endpoint.",
					},
					"transport": schema.StringAttribute{
						Optional:    true,
						Description: "Transport type, defaults to grpc.",
					},
					"tls": schema.BoolAttribute{
						Optional:    true,
						Description: "Reserved for TLS support.",
					},
				},
			},
			"startup_commands": schema.SingleNestedBlock{
				Description: "Reusable command refs to run after Create.",
				Attributes: map[string]schema.Attribute{
					"commands": schema.DynamicAttribute{
						Optional:    true,
						Description: "List of st2138_command resources, command objects, or command OID strings.",
					},
				},
			},
			"shutdown_commands": schema.SingleNestedBlock{
				Description: "Reusable command refs to run during Delete.",
				Attributes: map[string]schema.Attribute{
					"commands": schema.DynamicAttribute{
						Optional:    true,
						Description: "List of st2138_command resources, command objects, or command OID strings.",
					},
				},
			},
			// Legacy startup_command/shutdown_command blocks are intentionally omitted
			// from schema to match example usage.
		},
	}
}

func (r *deviceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if base, ok := req.ProviderData.(*clientpkg.Client); ok && base != nil {
		r.client = base.Clone()
	}
}

// configureClient sets endpoint and transport on the resource client from resource-level network attributes.
func (r *deviceResource) configureClient(network types.Object) diag.Diagnostics {
	var diags diag.Diagnostics
	if r.client == nil {
		r.client = &clientpkg.Client{}
	}
	if network.IsNull() || network.IsUnknown() {
		return diags
	}

	networkValues := network.Attributes()
	addressVal, okAddress := networkValues["address"].(types.String)
	portVal, okPort := networkValues["port"].(types.Int64)
	transportVal, okTransport := networkValues["transport"].(types.String)
	if !okAddress || !okPort || !okTransport {
		diags.AddError("invalid network configuration", "network must include address, port, and transport values")
		return diags
	}

	address := addressVal.ValueString()
	port := portVal.ValueInt64()
	endpoint := fmt.Sprintf("%s:%d", address, port)
	r.client.SetEndpoint(endpoint)

	transport := transportVal.ValueString()
	if transport != "" {
		r.client.Transport = transport
	} else {
		r.client.Transport = "grpc"
	}

	return diags
}

// parseParameters decodes the dynamic parameters attribute into map[oid]value.
// Supported shapes: {"counter": 1}, [{"counter": 1}], or [{...}, {...}] (merged).
func (r *deviceResource) parseParameters(params types.Dynamic) (map[string]attr.Value, diag.Diagnostics) {
	result := make(map[string]attr.Value)
	var diags diag.Diagnostics
	if params.IsNull() || params.IsUnknown() || params.UnderlyingValue() == nil {
		return result, diags
	}

	if directMap, err := r.attrMap(params.UnderlyingValue()); err == nil {
		for oid, v := range directMap {
			result[oid] = v
		}
		return result, diags
	}

	entries, err := r.attrSequence(params.UnderlyingValue())
	if err != nil {
		diags.AddError("invalid parameters value", "parameters must be an object/map or a list/tuple of objects")
		return result, diags
	}
	for _, entry := range entries {
		paramMap, mapErr := r.attrMap(entry)
		if mapErr != nil {
			diags.AddError("invalid parameter object", mapErr.Error())
			return result, diags
		}
		for oid, valAttr := range paramMap {
			result[oid] = valAttr
		}
	}

	return result, diags
}

// toFOID returns the fully-qualified OID form (leading "/") that Catena requires.
func toFOID(oid string) string {
	oid = strings.TrimSpace(oid)
	if oid == "" {
		return ""
	}
	return "/" + strings.TrimPrefix(oid, "/")
}

// changedParams returns entries in next that are new or differ from prev.
func changedParams(prev, next map[string]attr.Value) map[string]attr.Value {
	changed := make(map[string]attr.Value)
	for oid, v := range next {
		if old, ok := prev[oid]; ok && old.Equal(v) {
			continue
		}
		changed[oid] = v
	}
	return changed
}

// applySlotParams sets parameters on the device for the configured slot.
func (r *deviceResource) applySlotParams(ctx context.Context, slotNum uint32, oidValues map[string]attr.Value, diags *diag.Diagnostics) {
	snapshot, err := r.client.GetDeviceSnapshot(ctx, slotNum)
	if err != nil {
		diags.AddError("failed to read device model", fmt.Sprintf("DeviceRequest slot %d: %s", slotNum, err))
		return
	}
	missing := make([]string, 0)
	for oid := range oidValues {
		foid := toFOID(oid)
		if _, ok := snapshot.FullParameters[foid]; ok {
			continue
		}
		if _, ok := snapshot.FullParameters[strings.TrimPrefix(foid, "/")]; ok {
			continue
		}
		missing = append(missing, oid)
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		for _, oid := range missing {
			diags.AddAttributeError(path.Root("parameters"), "Unknown parameter",
				fmt.Sprintf("parameters.%q does not exist in the device model (slot %d)", oid, slotNum))
		}
		return
	}

	for key, rawValue := range oidValues {
		oid := toFOID(key)
		descriptor, err := r.client.GetParamDescriptor(ctx, slotNum, oid)
		if err != nil {
			diags.AddError("gRPC GetParam failed", fmt.Sprintf("slot %d oid %s: %s", slotNum, oid, err))
			return
		}
		protoValue, err := r.attrValueToProtoValue(rawValue, descriptor)
		if err != nil {
			diags.AddError("unsupported parameter value", fmt.Sprintf("slot %d oid %s: %s", slotNum, oid, err))
			return
		}
		if err := r.setRawValueWithRetry(ctx, slotNum, oid, protoValue); err != nil {
			diags.AddError("gRPC SetValue failed", fmt.Sprintf("slot %d oid %s: %s", slotNum, oid, err))
			return
		}
	}
}

type commandInvocation struct {
	OID                     string
	Value                   attr.Value
	StatusFoid              string
	StatusSuccessValue      string
	StatusSuccessComparator string
	TimeoutSeconds          int64
}

func (r *deviceResource) runCommandRefsBlock(ctx context.Context, slotNum uint32, block *commandRefsBlockModel, ignoreErrors bool) (string, error) {
	if block == nil || block.Commands.IsNull() || block.Commands.IsUnknown() || block.Commands.UnderlyingValue() == nil {
		return "", nil
	}

	entries, err := r.attrSequence(block.Commands.UnderlyingValue())
	if err != nil {
		return "", fmt.Errorf("commands must be a list or tuple: %w", err)
	}

	lastStatus := ""
	for _, entry := range entries {
		invocation, convErr := r.commandInvocationFromAttr(entry)
		if convErr != nil {
			if ignoreErrors {
				fmt.Fprintf(os.Stderr, "command conversion error: %v (ignored)\n", convErr)
				continue
			}
			return lastStatus, convErr
		}

		var protoValue *st2138pb.Value
		if invocation.Value != nil && !invocation.Value.IsNull() && !invocation.Value.IsUnknown() {
			protoValue, convErr = r.attrValueToProtoValue(invocation.Value, nil)
			if convErr != nil {
				if ignoreErrors {
					fmt.Fprintf(os.Stderr, "command value conversion error for %s: %v (ignored)\n", invocation.OID, convErr)
					continue
				}
				return lastStatus, fmt.Errorf("command %s value: %w", invocation.OID, convErr)
			}
		}

		if execErr := r.client.ExecuteCommand(ctx, slotNum, invocation.OID, protoValue); execErr != nil {
			if ignoreErrors {
				fmt.Fprintf(os.Stderr, "command %s: %v (ignored)\n", invocation.OID, execErr)
				continue
			}
			return lastStatus, fmt.Errorf("slot %d command %s: %w", slotNum, invocation.OID, execErr)
		}

		if invocation.StatusFoid != "" {
			if pollErr := r.pollStatusValue(ctx, slotNum, invocation.StatusFoid, invocation.StatusSuccessComparator, invocation.StatusSuccessValue, invocation.TimeoutSeconds); pollErr != nil {
				if ignoreErrors {
					fmt.Fprintf(os.Stderr, "command %s status poll failed: %v (ignored)\n", invocation.OID, pollErr)
					continue
				}
				return lastStatus, pollErr
			}
			if raw, getErr := r.client.GetRawValue(ctx, slotNum, invocation.StatusFoid); getErr == nil {
				lastStatus = protoValueToString(raw)
			}
		}
	}

	return lastStatus, nil
}

func (r *deviceResource) commandInvocationFromAttr(value attr.Value) (commandInvocation, error) {
	invocation := commandInvocation{
		StatusSuccessComparator: "eq",
		TimeoutSeconds:          5,
	}

	if v, ok := value.(types.String); ok {
		invocation.OID = toFOID(v.ValueString())
		if invocation.OID == "" {
			return invocation, fmt.Errorf("empty command oid")
		}
		return invocation, nil
	}

	fields, err := r.attrMap(value)
	if err != nil {
		return invocation, fmt.Errorf("command entry must be a string or object: %w", err)
	}

	if s, ok := attrToString(fields["command"]); ok && strings.TrimSpace(s) != "" {
		invocation.OID = toFOID(s)
	}

	if invocation.OID == "" {
		return invocation, fmt.Errorf("command object is missing command")
	}

	if v, ok := fields["value"]; ok {
		invocation.Value = v
	}

	if s, ok := attrToString(fields["status_foid"]); ok {
		invocation.StatusFoid = toFOID(s)
	}
	if s, ok := attrToString(fields["status_success_value"]); ok {
		invocation.StatusSuccessValue = s
	}
	if s, ok := attrToString(fields["status_success_comparator"]); ok && s != "" {
		invocation.StatusSuccessComparator = s
	}
	if n, ok := attrToInt64(fields["timeout_seconds"]); ok && n > 0 {
		invocation.TimeoutSeconds = n
	}

	return invocation, nil
}

func attrToString(value attr.Value) (string, bool) {
	if value == nil {
		return "", false
	}
	switch v := value.(type) {
	case types.String:
		if v.IsNull() || v.IsUnknown() {
			return "", false
		}
		return v.ValueString(), true
	case types.Dynamic:
		if v.IsNull() || v.IsUnknown() || v.UnderlyingValue() == nil {
			return "", false
		}
		return attrToString(v.UnderlyingValue())
	default:
		return "", false
	}
}

func attrToInt64(value attr.Value) (int64, bool) {
	if value == nil {
		return 0, false
	}
	switch v := value.(type) {
	case types.Int64:
		if v.IsNull() || v.IsUnknown() {
			return 0, false
		}
		return v.ValueInt64(), true
	case types.Number:
		if v.IsNull() || v.IsUnknown() || v.ValueBigFloat() == nil {
			return 0, false
		}
		n, _ := v.ValueBigFloat().Int64()
		return n, true
	case types.Dynamic:
		if v.IsNull() || v.IsUnknown() || v.UnderlyingValue() == nil {
			return 0, false
		}
		return attrToInt64(v.UnderlyingValue())
	default:
		return 0, false
	}
}

// Legacy single-block helpers (startup_command/shutdown_command) were removed
// from active schema because current examples use startup_commands/shutdown_commands.

// pollStatusValue polls the given foid every 500ms until the comparator condition is met or timeout.
func (r *deviceResource) pollStatusValue(ctx context.Context, slotNum uint32, foid, comparator, target string, timeoutSecs int64) error {
	deadline := time.Now().Add(time.Duration(timeoutSecs) * time.Second)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for slot %d %s %s %q", slotNum, foid, comparator, target)
		}
		val, err := r.client.GetRawValue(ctx, slotNum, foid)
		if err != nil {
			return fmt.Errorf("GetValue slot %d %s: %w", slotNum, foid, err)
		}
		current := protoValueToString(val)
		if compareStatus(current, comparator, target) {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// protoValueToString converts a proto Value to a comparable string.
func protoValueToString(v *st2138pb.Value) string {
	if v == nil {
		return ""
	}
	switch k := v.GetKind().(type) {
	case *st2138pb.Value_StringValue:
		return k.StringValue
	case *st2138pb.Value_Int32Value:
		return strconv.Itoa(int(k.Int32Value))
	case *st2138pb.Value_Float32Value:
		return strconv.FormatFloat(float64(k.Float32Value), 'f', -1, 32)
	default:
		return ""
	}
}

// compareStatus compares two numeric-or-string values using the given operator.
func compareStatus(current, op, target string) bool {
	// Try numeric comparison first.
	cf, cerr := strconv.ParseFloat(current, 64)
	tf, terr := strconv.ParseFloat(target, 64)
	if cerr == nil && terr == nil {
		switch op {
		case "eq":
			return cf == tf
		case "ne":
			return cf != tf
		case "gt":
			return cf > tf
		case "lt":
			return cf < tf
		case "ge":
			return cf >= tf
		case "le":
			return cf <= tf
		}
	}
	// Fallback to string comparison.
	switch op {
	case "eq":
		return current == target
	case "ne":
		return current != target
	}
	return false
}

func (r *deviceResource) attrSequence(value attr.Value) ([]attr.Value, error) {
	switch v := value.(type) {
	case types.Dynamic:
		if v.IsNull() || v.IsUnknown() || v.UnderlyingValue() == nil {
			return nil, fmt.Errorf("dynamic sequence is null or unknown")
		}
		return r.attrSequence(v.UnderlyingValue())
	case types.List:
		return v.Elements(), nil
	case types.Tuple:
		return v.Elements(), nil
	default:
		return nil, fmt.Errorf("expected list or tuple, got %T", value)
	}
}

func (r *deviceResource) attrMap(value attr.Value) (map[string]attr.Value, error) {
	switch v := value.(type) {
	case types.Dynamic:
		if v.IsNull() || v.IsUnknown() || v.UnderlyingValue() == nil {
			return nil, fmt.Errorf("dynamic map is null or unknown")
		}
		return r.attrMap(v.UnderlyingValue())
	case types.Map:
		return v.Elements(), nil
	case types.Object:
		return v.Attributes(), nil
	default:
		return nil, fmt.Errorf("expected map or object, got %T", value)
	}
}

func (r *deviceResource) attrValueToProtoValue(value attr.Value, descriptor *st2138pb.Param) (*st2138pb.Value, error) {
	if dynamic, ok := value.(types.Dynamic); ok {
		if dynamic.IsNull() || dynamic.IsUnknown() || dynamic.UnderlyingValue() == nil {
			return &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}}, nil
		}
		return r.attrValueToProtoValue(dynamic.UnderlyingValue(), descriptor)
	}

	if descriptor != nil {
		switch descriptor.GetType() {
		case st2138pb.ParamType_BINARY:
			return r.dataPayloadToProtoValue(value)
		case st2138pb.ParamType_STRUCT_VARIANT:
			return r.structVariantToProtoValue(value, descriptor)
		case st2138pb.ParamType_STRUCT_VARIANT_ARRAY:
			return r.structVariantArrayToProtoValue(value, descriptor)
		}
	}

	switch v := value.(type) {
	case types.String:
		if descriptor != nil {
			switch descriptor.GetType() {
			case st2138pb.ParamType_INT32:
				if iv, err := strconv.ParseInt(v.ValueString(), 10, 32); err == nil {
					return &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: int32(iv)}}, nil
				}
			case st2138pb.ParamType_FLOAT32:
				if fv, err := strconv.ParseFloat(v.ValueString(), 64); err == nil {
					return &st2138pb.Value{Kind: &st2138pb.Value_Float32Value{Float32Value: float32(fv)}}, nil
				}
			}
		}
		return &st2138pb.Value{Kind: &st2138pb.Value_StringValue{StringValue: v.ValueString()}}, nil
	case types.Bool:
		s := "false"
		if v.ValueBool() {
			s = "true"
		}
		return &st2138pb.Value{Kind: &st2138pb.Value_StringValue{StringValue: s}}, nil
	case types.Number:
		return r.numberToProtoValue(v, descriptor)
	case types.List:
		return r.sequenceToProtoValue(v.Elements(), descriptor)
	case types.Tuple:
		return r.sequenceToProtoValue(v.Elements(), descriptor)
	case types.Map:
		return r.objectToProtoValue(v.Elements(), descriptor)
	case types.Object:
		return r.objectToProtoValue(v.Attributes(), descriptor)
	default:
		return nil, fmt.Errorf("unsupported value type %T", value)
	}
}

func (r *deviceResource) dataPayloadToProtoValue(value attr.Value) (*st2138pb.Value, error) {
	fields, err := r.attrMap(value)
	if err != nil {
		return nil, fmt.Errorf("data payload must be an object: %w", err)
	}
	if nested, ok := fields["data_payload"]; ok && len(fields) == 1 {
		fields, err = r.attrMap(nested)
		if err != nil {
			return nil, fmt.Errorf("data_payload must be an object: %w", err)
		}
	}

	payload := &st2138pb.DataPayload{}
	if rawMetadata, ok := fields["metadata"]; ok {
		metadata, err := r.attrMap(rawMetadata)
		if err != nil {
			return nil, fmt.Errorf("data payload metadata must be an object: %w", err)
		}
		payload.Metadata = make(map[string]string, len(metadata))
		for key, rawValue := range metadata {
			metadataValue, err := stringAttrValue(rawValue)
			if err != nil {
				return nil, fmt.Errorf("data payload metadata %q: %w", key, err)
			}
			payload.Metadata[key] = metadataValue
		}
	}

	if rawEncoding, ok := fields["payload_encoding"]; ok {
		encoding, err := stringAttrValue(rawEncoding)
		if err != nil {
			return nil, fmt.Errorf("data payload encoding: %w", err)
		}
		encodingValue, ok := st2138pb.DataPayload_PayloadEncoding_value[strings.ToUpper(encoding)]
		if !ok {
			return nil, fmt.Errorf("unsupported data payload encoding %q", encoding)
		}
		payload.PayloadEncoding = st2138pb.DataPayload_PayloadEncoding(encodingValue)
	}

	payloadSources := 0
	for _, source := range []string{"payload", "payload_file", "url"} {
		if _, ok := fields[source]; ok {
			payloadSources++
		}
	}
	if payloadSources != 1 {
		return nil, fmt.Errorf("data payload must set exactly one of payload, payload_file, or url")
	}

	var payloadBytes []byte
	if rawPayload, ok := fields["payload"]; ok {
		encoded, err := stringAttrValue(rawPayload)
		if err != nil {
			return nil, fmt.Errorf("data payload payload must be a base64 string: %w", err)
		}
		payloadBytes, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("data payload payload must be valid base64: %w", err)
		}
		payload.Kind = &st2138pb.DataPayload_Payload{Payload: payloadBytes}
	}
	if rawFile, ok := fields["payload_file"]; ok {
		filePath, err := stringAttrValue(rawFile)
		if err != nil {
			return nil, fmt.Errorf("data payload payload_file must be a string: %w", err)
		}
		payloadBytes, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read data payload file %q: %w", filePath, err)
		}
		payload.Kind = &st2138pb.DataPayload_Payload{Payload: payloadBytes}
	}
	if rawURL, ok := fields["url"]; ok {
		url, err := stringAttrValue(rawURL)
		if err != nil {
			return nil, fmt.Errorf("data payload url must be a string: %w", err)
		}
		payload.Kind = &st2138pb.DataPayload_Url{Url: url}
	}

	if rawDigest, ok := fields["digest"]; ok {
		digest, err := stringAttrValue(rawDigest)
		if err != nil {
			return nil, fmt.Errorf("data payload digest must be a base64 string: %w", err)
		}
		if digest != "" {
			payload.Digest, err = base64.StdEncoding.DecodeString(digest)
			if err != nil {
				return nil, fmt.Errorf("data payload digest must be valid base64: %w", err)
			}
		}
	}
	if len(payload.Digest) == 0 && payload.GetKind() != nil {
		if _, isBytes := payload.GetKind().(*st2138pb.DataPayload_Payload); isBytes {
			checksum := sha256.Sum256(payloadBytes)
			payload.Digest = checksum[:]
		}
	}

	return &st2138pb.Value{Kind: &st2138pb.Value_DataPayload{DataPayload: payload}}, nil
}

func stringAttrValue(value attr.Value) (string, error) {
	if dynamic, ok := value.(types.Dynamic); ok {
		if dynamic.IsNull() || dynamic.IsUnknown() || dynamic.UnderlyingValue() == nil {
			return "", fmt.Errorf("value is null or unknown")
		}
		return stringAttrValue(dynamic.UnderlyingValue())
	}
	stringValue, ok := value.(types.String)
	if !ok || stringValue.IsNull() || stringValue.IsUnknown() {
		return "", fmt.Errorf("expected string, got %T", value)
	}
	return stringValue.ValueString(), nil
}

func (r *deviceResource) structVariantToProtoValue(value attr.Value, descriptor *st2138pb.Param) (*st2138pb.Value, error) {
	fields, err := r.attrMap(value)
	if err != nil {
		return nil, fmt.Errorf("struct variant must be an object: %w", err)
	}
	if _, hasVariantType := fields["struct_variant_type"]; !hasVariantType && len(fields) == 1 {
		if nested, ok := fields["nested_struct"]; ok {
			fields, err = r.attrMap(nested)
			if err != nil {
				return nil, fmt.Errorf("nested struct variant must be an object: %w", err)
			}
		}
	}

	rawVariantType, ok := fields["struct_variant_type"]
	if !ok {
		return nil, fmt.Errorf("struct variant is missing struct_variant_type")
	}
	variantTypeValue, err := r.attrValueToProtoValue(rawVariantType, nil)
	if err != nil {
		return nil, fmt.Errorf("struct variant type: %w", err)
	}
	variantType, ok := variantTypeValue.GetKind().(*st2138pb.Value_StringValue)
	if !ok {
		return nil, fmt.Errorf("struct_variant_type must be a string")
	}

	rawValue, ok := fields["value"]
	if !ok {
		return nil, fmt.Errorf("struct variant is missing value")
	}
	var valueDescriptor *st2138pb.Param
	if descriptor != nil {
		valueDescriptor = descriptor.GetParams()[variantType.StringValue]
	}
	protoValue, err := r.attrValueToProtoValue(rawValue, valueDescriptor)
	if err != nil {
		return nil, fmt.Errorf("struct variant value: %w", err)
	}

	return &st2138pb.Value{Kind: &st2138pb.Value_StructVariantValue{
		StructVariantValue: &st2138pb.StructVariantValue{
			StructVariantType: variantType.StringValue,
			Value:             protoValue,
		},
	}}, nil
}

func (r *deviceResource) structVariantArrayToProtoValue(value attr.Value, descriptor *st2138pb.Param) (*st2138pb.Value, error) {
	elements, err := r.attrSequence(value)
	if err != nil {
		return nil, fmt.Errorf("struct variant array must be a list or tuple: %w", err)
	}

	variants := make([]*st2138pb.StructVariantValue, 0, len(elements))
	for index, element := range elements {
		converted, err := r.structVariantToProtoValue(element, descriptor)
		if err != nil {
			return nil, fmt.Errorf("variant %d: %w", index, err)
		}
		variants = append(variants, converted.GetStructVariantValue())
	}

	return &st2138pb.Value{Kind: &st2138pb.Value_StructVariantArrayValues{
		StructVariantArrayValues: &st2138pb.StructVariantList{StructVariants: variants},
	}}, nil
}

func (r *deviceResource) numberToProtoValue(v types.Number, descriptor *st2138pb.Param) (*st2138pb.Value, error) {
	bf := v.ValueBigFloat()
	if bf == nil {
		return &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}}, nil
	}
	if descriptor != nil && descriptor.GetType() == st2138pb.ParamType_FLOAT32 {
		fv, _ := bf.Float64()
		return &st2138pb.Value{Kind: &st2138pb.Value_Float32Value{Float32Value: float32(fv)}}, nil
	}
	if iv, acc := bf.Int64(); acc == big.Exact && iv >= math.MinInt32 && iv <= math.MaxInt32 {
		return &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: int32(iv)}}, nil
	}
	fv, _ := bf.Float64()
	return &st2138pb.Value{Kind: &st2138pb.Value_Float32Value{Float32Value: float32(fv)}}, nil
}

func (r *deviceResource) objectToProtoValue(fields map[string]attr.Value, descriptor *st2138pb.Param) (*st2138pb.Value, error) {
	result := make(map[string]*st2138pb.Value, len(fields))
	for key, field := range fields {
		var childDescriptor *st2138pb.Param
		if descriptor != nil && descriptor.GetParams() != nil {
			childDescriptor = descriptor.GetParams()[key]
		}
		converted, err := r.attrValueToProtoValue(field, childDescriptor)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", key, err)
		}
		result[key] = converted
	}
	return &st2138pb.Value{Kind: &st2138pb.Value_StructValue{StructValue: &st2138pb.StructValue{Fields: result}}}, nil
}

func (r *deviceResource) sequenceToProtoValue(elements []attr.Value, descriptor *st2138pb.Param) (*st2138pb.Value, error) {
	if len(elements) == 0 {
		switch {
		case descriptor != nil && descriptor.GetType() == st2138pb.ParamType_INT32_ARRAY:
			return &st2138pb.Value{Kind: &st2138pb.Value_Int32ArrayValues{Int32ArrayValues: &st2138pb.Int32List{Ints: []int32{}}}}, nil
		case descriptor != nil && descriptor.GetType() == st2138pb.ParamType_FLOAT32_ARRAY:
			return &st2138pb.Value{Kind: &st2138pb.Value_Float32ArrayValues{Float32ArrayValues: &st2138pb.Float32List{Floats: []float32{}}}}, nil
		case descriptor != nil && descriptor.GetType() == st2138pb.ParamType_STRUCT_ARRAY:
			return &st2138pb.Value{Kind: &st2138pb.Value_StructArrayValues{StructArrayValues: &st2138pb.StructList{StructValues: []*st2138pb.StructValue{}}}}, nil
		default:
			return &st2138pb.Value{Kind: &st2138pb.Value_StringArrayValues{StringArrayValues: &st2138pb.StringList{Strings: []string{}}}}, nil
		}
	}

	allStrings := true
	allNumbers := true
	allObjects := true
	allInts := true

	stringVals := make([]string, 0, len(elements))
	intVals := make([]int32, 0, len(elements))
	floatVals := make([]float32, 0, len(elements))
	structVals := make([]*st2138pb.StructValue, 0, len(elements))

	for _, elem := range elements {
		switch v := elem.(type) {
		case types.Dynamic:
			if v.IsNull() || v.IsUnknown() || v.UnderlyingValue() == nil {
				allStrings = false
				allNumbers = false
				allObjects = false
				allInts = false
				continue
			}
			converted, err := r.sequenceToProtoValue([]attr.Value{v.UnderlyingValue()}, descriptor)
			if err != nil {
				return nil, err
			}
			switch kind := converted.GetKind().(type) {
			case *st2138pb.Value_StringArrayValues:
				stringVals = append(stringVals, kind.StringArrayValues.GetStrings()[0])
				allNumbers = false
				allObjects = false
			case *st2138pb.Value_Int32ArrayValues:
				intVals = append(intVals, kind.Int32ArrayValues.GetInts()[0])
				floatVals = append(floatVals, float32(kind.Int32ArrayValues.GetInts()[0]))
				allStrings = false
				allObjects = false
			case *st2138pb.Value_Float32ArrayValues:
				allInts = false
				floatVals = append(floatVals, kind.Float32ArrayValues.GetFloats()[0])
				allStrings = false
				allObjects = false
			case *st2138pb.Value_StructArrayValues:
				structVals = append(structVals, kind.StructArrayValues.GetStructValues()[0])
				allStrings = false
				allNumbers = false
				allInts = false
			default:
				return nil, fmt.Errorf("unsupported array element type %T", elem)
			}
		case types.String:
			stringVals = append(stringVals, v.ValueString())
			allNumbers = false
			allObjects = false
			allInts = false
		case types.Number:
			bf := v.ValueBigFloat()
			if bf == nil {
				return nil, fmt.Errorf("null number in array")
			}
			fv, _ := bf.Float64()
			floatVals = append(floatVals, float32(fv))
			if descriptor != nil && descriptor.GetType() == st2138pb.ParamType_FLOAT32_ARRAY {
				allInts = false
			} else if iv, acc := bf.Int64(); acc == big.Exact && iv >= math.MinInt32 && iv <= math.MaxInt32 {
				intVals = append(intVals, int32(iv))
			} else {
				allInts = false
			}
			allStrings = false
			allObjects = false
		case types.Map:
			converted, err := r.objectToProtoValue(v.Elements(), descriptor)
			if err != nil {
				return nil, err
			}
			structVals = append(structVals, converted.GetStructValue())
			allStrings = false
			allNumbers = false
			allInts = false
		case types.Object:
			converted, err := r.objectToProtoValue(v.Attributes(), descriptor)
			if err != nil {
				return nil, err
			}
			structVals = append(structVals, converted.GetStructValue())
			allStrings = false
			allNumbers = false
			allInts = false
		default:
			allStrings = false
			allNumbers = false
			allObjects = false
			allInts = false
		}
	}

	if allStrings {
		return &st2138pb.Value{Kind: &st2138pb.Value_StringArrayValues{StringArrayValues: &st2138pb.StringList{Strings: stringVals}}}, nil
	}
	if allNumbers {
		if allInts {
			return &st2138pb.Value{Kind: &st2138pb.Value_Int32ArrayValues{Int32ArrayValues: &st2138pb.Int32List{Ints: intVals}}}, nil
		}
		return &st2138pb.Value{Kind: &st2138pb.Value_Float32ArrayValues{Float32ArrayValues: &st2138pb.Float32List{Floats: floatVals}}}, nil
	}
	if allObjects {
		return &st2138pb.Value{Kind: &st2138pb.Value_StructArrayValues{StructArrayValues: &st2138pb.StructList{StructValues: structVals}}}, nil
	}

	return nil, fmt.Errorf("mixed-type arrays are not supported")
}

// buildSnapshotMaps reads the slot snapshot and returns computed output maps.
func (r *deviceResource) buildSnapshotMaps(ctx context.Context, slotNum uint32) (types.Map, types.Map, types.Map, error) {
	snapshot, err := r.client.GetDeviceSnapshot(ctx, slotNum)
	if err != nil {
		return types.MapNull(types.StringType), types.MapNull(types.StringType), types.MapNull(types.StringType), fmt.Errorf("DeviceRequest slot %d: %w", slotNum, err)
	}

	paramElems := make(map[string]attr.Value, len(snapshot.Parameters))
	for foid, value := range snapshot.Parameters {
		paramElems[foid] = types.StringValue(value)
	}
	paramsMap, mapDiags := types.MapValue(types.StringType, paramElems)
	if mapDiags.HasError() {
		return types.MapNull(types.StringType), types.MapNull(types.StringType), types.MapNull(types.StringType), fmt.Errorf("building parameters_out map")
	}

	fullParamElems := make(map[string]attr.Value, len(snapshot.FullParameters))
	for foid, value := range snapshot.FullParameters {
		fullParamElems[foid] = types.StringValue(value)
	}
	fullParamsMap, fullMapDiags := types.MapValue(types.StringType, fullParamElems)
	if fullMapDiags.HasError() {
		return types.MapNull(types.StringType), types.MapNull(types.StringType), types.MapNull(types.StringType), fmt.Errorf("building full_parameters_out map")
	}

	commandElems := make(map[string]attr.Value, len(snapshot.Commands))
	for foid, value := range snapshot.Commands {
		commandElems[foid] = types.StringValue(value)
	}
	commandsMap, commandMapDiags := types.MapValue(types.StringType, commandElems)
	if commandMapDiags.HasError() {
		return types.MapNull(types.StringType), types.MapNull(types.StringType), types.MapNull(types.StringType), fmt.Errorf("building commands_out map")
	}

	return paramsMap, fullParamsMap, commandsMap, nil
}

func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.configureClient(plan.Network)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Name.IsNull() || plan.Name.ValueString() == "" {
		plan.Name = types.StringValue("device")
	}
	plan.ID = types.StringValue("catena-" + plan.Name.ValueString())

	if plan.SlotID.IsNull() || plan.SlotID.IsUnknown() {
		resp.Diagnostics.AddError("invalid slot", "slot must be set for st2138_device")
		return
	}
	slotNum := uint32(plan.SlotID.ValueInt64())

	// Apply configured parameters to the device.
	paramValues, parseDiags := r.parseParameters(plan.Parameters)
	resp.Diagnostics.Append(parseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(paramValues) > 0 {
		r.applySlotParams(ctx, slotNum, paramValues, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	statusValue, err := r.runCommandRefsBlock(ctx, slotNum, plan.StartupCommands, false)
	if err != nil {
		resp.Diagnostics.AddError("startup_commands failed", err.Error())
		return
	}
	plan.StatusValue = types.StringValue(statusValue)

	paramsMap, fullParamsMap, commandsMap, err := r.buildSnapshotMaps(ctx, slotNum)
	if err != nil {
		resp.Diagnostics.AddError("failed to read device model", err.Error())
		return
	}
	plan.ParametersOut = paramsMap
	plan.FullParametersOut = fullParamsMap
	plan.CommandsOut = commandsMap
	if plan.StatusValue.IsNull() {
		plan.StatusValue = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.configureClient(state.Network)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SlotID.IsNull() || state.SlotID.IsUnknown() {
		resp.Diagnostics.AddError("invalid slot", "slot must be set for st2138_device")
		return
	}
	slotNum := uint32(state.SlotID.ValueInt64())

	_, parseDiags := r.parseParameters(state.Parameters)
	resp.Diagnostics.Append(parseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	paramsMap, fullParamsMap, commandsMap, err := r.buildSnapshotMaps(ctx, slotNum)
	if err != nil {
		resp.Diagnostics.AddError("failed to read device model", err.Error())
		return
	}
	state.ParametersOut = paramsMap
	state.FullParametersOut = fullParamsMap
	state.CommandsOut = commandsMap
	if state.StatusValue.IsNull() {
		state.StatusValue = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.configureClient(plan.Network)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve ID from previous state.
	var prev deviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.ID.IsNull() || plan.ID.ValueString() == "" {
		plan.ID = prev.ID
	}

	if plan.SlotID.IsNull() || plan.SlotID.IsUnknown() {
		resp.Diagnostics.AddError("invalid slot", "slot must be set for st2138_device")
		return
	}
	slotNum := uint32(plan.SlotID.ValueInt64())

	planParams, parseDiags := r.parseParameters(plan.Parameters)
	resp.Diagnostics.Append(parseDiags...)
	prevParams, prevDiags := r.parseParameters(prev.Parameters)
	resp.Diagnostics.Append(prevDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if changed := changedParams(prevParams, planParams); len(changed) > 0 {
		r.applySlotParams(ctx, slotNum, changed, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	paramsMap, fullParamsMap, commandsMap, err := r.buildSnapshotMaps(ctx, slotNum)
	if err != nil {
		resp.Diagnostics.AddError("failed to read device model", err.Error())
		return
	}
	plan.ParametersOut = paramsMap
	plan.FullParametersOut = fullParamsMap
	plan.CommandsOut = commandsMap
	if plan.StatusValue.IsUnknown() {
		plan.StatusValue = prev.StatusValue
	}
	if plan.StatusValue.IsNull() || plan.StatusValue.IsUnknown() {
		plan.StatusValue = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ShutdownCommands != nil {
		resp.Diagnostics.Append(r.configureClient(state.Network)...)
		if resp.Diagnostics.HasError() {
			return
		}
		slotNum := uint32(state.SlotID.ValueInt64())
		if _, err := r.runCommandRefsBlock(ctx, slotNum, state.ShutdownCommands, true); err != nil {
			resp.Diagnostics.AddWarning("shutdown_commands warning", fmt.Sprintf("shutdown command encountered issues: %v (continuing with destroy)", err))
		}
	}
	// Remove from Terraform state only; no further gRPC calls needed.
}

// Helpers

func (r *deviceResource) resolveDevicesDir() string {
	candidates := []string{
		"../opentofu/exe", "../../opentofu/exe", "/workspace/opentofu/exe", "./opentofu/exe",
		"../exe", "../../exe", "/workspace/exe", "./exe",
	}
	if r.client != nil && r.client.DevicesDir != "" {
		candidates = []string{r.client.DevicesDir}
	}
	for _, c := range candidates {
		if r.pathExists(c) {
			return c
		}
	}
	return "."
}

func (r *deviceResource) pathExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func (r *deviceResource) commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (r *deviceResource) selectHostPortForInternal(ports []string, internal int) int {
	for _, p := range ports {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		if idx := strings.IndexByte(s, '/'); idx >= 0 {
			s = s[:idx]
		}
		parts := strings.Split(s, ":")
		if len(parts) < 2 {
			continue
		}
		cpart := parts[len(parts)-1]
		hpart := parts[len(parts)-2]
		if cp, err := strconv.Atoi(cpart); err == nil && cp == internal {
			if hp, err2 := strconv.Atoi(hpart); err2 == nil {
				return hp
			}
		}
	}
	return 0
}

func (r *deviceResource) dockerHost() string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "host.docker.internal"
	}
	return "localhost"
}

func (r *deviceResource) dockerExec(ctx context.Context, name, command string) error {
	cmd := exec.CommandContext(ctx, "docker", "exec", name, "/bin/sh", "-lc", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker exec error: %v\n%s", err, string(out))
	}
	return nil
}

func (r *deviceResource) getContainerID(name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Names}}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) >= 2 && parts[1] == name {
			return parts[0]
		}
	}
	return ""
}

func (r *deviceResource) parseValueString(s string) any {
	if s == "" {
		return ""
	}
	if strings.EqualFold(s, "true") {
		return true
	}
	if strings.EqualFold(s, "false") {
		return false
	}
	if iv, err := strconv.ParseInt(s, 10, 32); err == nil {
		return float64(iv)
	}
	if fv, err := strconv.ParseFloat(s, 64); err == nil {
		return fv
	}
	return s
}

// Retries are handled by the client's SetValue interceptor.
func (r *deviceResource) setStringValueWithRetry(ctx context.Context, slot uint32, oid, value string) error {
	return r.client.SetStringValue(ctx, slot, oid, value)
}

func (r *deviceResource) setNumberValueWithRetry(ctx context.Context, slot uint32, oid string, n float64) error {
	return r.client.SetNumberValue(ctx, slot, oid, n)
}

func (r *deviceResource) setRawValueWithRetry(ctx context.Context, slot uint32, oid string, value *st2138pb.Value) error {
	return r.client.SetRawValue(ctx, slot, oid, value)
}
