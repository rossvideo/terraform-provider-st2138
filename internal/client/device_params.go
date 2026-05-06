package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

type DeviceSnapshot struct {
	Parameters     map[string]string
	FullParameters map[string]string
	Commands       map[string]string
}

// GetDeviceSnapshot calls DeviceRequest for the given slot and collects
// writable parameters, all parameters, and commands from the streamed response.
func (c *Client) GetDeviceSnapshot(ctx context.Context, slot uint32) (*DeviceSnapshot, error) {
	if err := c.ensureConn(ctx); err != nil {
		return nil, err
	}

	stream, err := c.rpcClient.DeviceRequest(ctx, &st2138pb.DeviceRequestPayload{
		Slot:        slot,
		DetailLevel: st2138pb.Device_FULL,
	})
	if err != nil {
		return nil, err
	}

	result := &DeviceSnapshot{
		Parameters:     make(map[string]string),
		FullParameters: make(map[string]string),
		Commands:       make(map[string]string),
	}

	for {
		comp, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return result, recvErr
		}

		if dev := comp.GetDevice(); dev != nil {
			for foid, param := range dev.GetParams() {
				value := stringifyValue(param.GetValue())
				result.FullParameters[foid] = value
				if !param.GetReadOnly() {
					result.Parameters[foid] = value
				}
			}
			for foid, cmd := range dev.GetCommands() {
				result.Commands[foid] = stringifyValue(cmd.GetValue())
			}
		}

		if cp := comp.GetParam(); cp != nil {
			value := stringifyValue(cp.GetParam().GetValue())
			result.FullParameters[cp.Oid] = value
			if !cp.GetParam().GetReadOnly() {
				result.Parameters[cp.Oid] = value
			}
		}

		if cc := comp.GetCommand(); cc != nil {
			result.Commands[cc.Oid] = stringifyValue(cc.GetCommand().GetValue())
		}
	}

	return result, nil
}

// GetDeviceParams calls DeviceRequest for the given slot and collects all
// parameter OIDs and their current values from the streamed response.
// Returns a map of fully-qualified OID (foid) -> stringified value,
// e.g. {"counter": "1", "label": "device-a"}.
func (c *Client) GetDeviceParams(ctx context.Context, slot uint32) (map[string]string, error) {
	snapshot, err := c.GetDeviceSnapshot(ctx, slot)
	if err != nil {
		return nil, err
	}
	return snapshot.Parameters, nil
}

func stringifyValue(v *st2138pb.Value) string {
	if v == nil {
		return ""
	}

	switch kind := v.GetKind().(type) {
	case *st2138pb.Value_StringValue:
		return kind.StringValue
	case *st2138pb.Value_Int32Value:
		return fmt.Sprintf("%d", kind.Int32Value)
	case *st2138pb.Value_Float32Value:
		return fmt.Sprintf("%g", kind.Float32Value)
	case *st2138pb.Value_EmptyValue:
		return ""
	case *st2138pb.Value_StringArrayValues:
		return mustJSONString(kind.StringArrayValues.GetStrings())
	case *st2138pb.Value_Int32ArrayValues:
		return mustJSONString(kind.Int32ArrayValues.GetInts())
	case *st2138pb.Value_Float32ArrayValues:
		return mustJSONString(kind.Float32ArrayValues.GetFloats())
	case *st2138pb.Value_StructValue:
		return stringifyStructValue(kind.StructValue)
	case *st2138pb.Value_StructArrayValues:
		items := make([]map[string]any, 0, len(kind.StructArrayValues.GetStructValues()))
		for _, item := range kind.StructArrayValues.GetStructValues() {
			items = append(items, structValueToMap(item))
		}
		return mustJSONString(items)
	case *st2138pb.Value_StructVariantValue:
		return mustJSONString(map[string]any{
			"struct_variant_type": kind.StructVariantValue.GetStructVariantType(),
			"value":               valueToAny(kind.StructVariantValue.GetValue()),
		})
	case *st2138pb.Value_StructVariantArrayValues:
		items := make([]map[string]any, 0, len(kind.StructVariantArrayValues.GetStructVariants()))
		for _, item := range kind.StructVariantArrayValues.GetStructVariants() {
			items = append(items, map[string]any{
				"struct_variant_type": item.GetStructVariantType(),
				"value":               valueToAny(item.GetValue()),
			})
		}
		return mustJSONString(items)
	case *st2138pb.Value_DataPayload:
		return mustJSONString(map[string]any{
			"metadata":         kind.DataPayload.GetMetadata(),
			"payload_encoding": kind.DataPayload.GetPayloadEncoding().String(),
			"url":              kind.DataPayload.GetUrl(),
		})
	default:
		return ""
	}
}

func stringifyStructValue(v *st2138pb.StructValue) string {
	return mustJSONString(structValueToMap(v))
}

func structValueToMap(v *st2138pb.StructValue) map[string]any {
	if v == nil {
		return map[string]any{}
	}
	result := make(map[string]any, len(v.GetFields()))
	keys := make([]string, 0, len(v.GetFields()))
	for key := range v.GetFields() {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result[key] = valueToAny(v.GetFields()[key])
	}
	return result
}

func valueToAny(v *st2138pb.Value) any {
	if v == nil {
		return ""
	}
	switch kind := v.GetKind().(type) {
	case *st2138pb.Value_StringValue:
		return kind.StringValue
	case *st2138pb.Value_Int32Value:
		return kind.Int32Value
	case *st2138pb.Value_Float32Value:
		return kind.Float32Value
	case *st2138pb.Value_EmptyValue:
		return ""
	case *st2138pb.Value_StringArrayValues:
		return kind.StringArrayValues.GetStrings()
	case *st2138pb.Value_Int32ArrayValues:
		return kind.Int32ArrayValues.GetInts()
	case *st2138pb.Value_Float32ArrayValues:
		return kind.Float32ArrayValues.GetFloats()
	case *st2138pb.Value_StructValue:
		return structValueToMap(kind.StructValue)
	case *st2138pb.Value_StructArrayValues:
		items := make([]map[string]any, 0, len(kind.StructArrayValues.GetStructValues()))
		for _, item := range kind.StructArrayValues.GetStructValues() {
			items = append(items, structValueToMap(item))
		}
		return items
	case *st2138pb.Value_StructVariantValue:
		return map[string]any{
			"struct_variant_type": kind.StructVariantValue.GetStructVariantType(),
			"value":               valueToAny(kind.StructVariantValue.GetValue()),
		}
	case *st2138pb.Value_StructVariantArrayValues:
		items := make([]map[string]any, 0, len(kind.StructVariantArrayValues.GetStructVariants()))
		for _, item := range kind.StructVariantArrayValues.GetStructVariants() {
			items = append(items, map[string]any{
				"struct_variant_type": item.GetStructVariantType(),
				"value":               valueToAny(item.GetValue()),
			})
		}
		return items
	case *st2138pb.Value_DataPayload:
		return map[string]any{
			"metadata":         kind.DataPayload.GetMetadata(),
			"payload_encoding": kind.DataPayload.GetPayloadEncoding().String(),
			"url":              kind.DataPayload.GetUrl(),
		}
	default:
		return ""
	}
}

func mustJSONString(v any) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bytes)
}
