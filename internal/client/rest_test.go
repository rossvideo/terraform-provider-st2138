package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestClient_RESTLifecycleRequests(t *testing.T) {
	device := &st2138pb.Device{
		Params: map[string]*st2138pb.Param{
			"counter": {Value: &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: 2}}},
			"locked":  {ReadOnly: true, Value: &st2138pb.Value{Kind: &st2138pb.Value_StringValue{StringValue: "fixed"}}},
		},
		Commands: map[string]*st2138pb.Param{
			"reset": {Value: &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}}},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/st2138-api/v1/1":
			if got := r.Header.Get("Detail-Level"); got != "FULL" {
				t.Errorf("Detail-Level = %q, want FULL", got)
			}
			writeRESTProtoResponse(t, w, http.StatusOK, device)
		case r.Method == http.MethodGet && r.URL.Path == "/st2138-api/v1/1/param/nested/counter":
			writeRESTProtoResponse(t, w, http.StatusOK, &st2138pb.DeviceComponent_ComponentParam{
				Oid:   "nested/counter",
				Param: &st2138pb.Param{Type: st2138pb.ParamType_INT32},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/st2138-api/v1/1/value/nested/counter":
			writeRESTProtoResponse(t, w, http.StatusOK, &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: 7}})
		case r.Method == http.MethodGet && r.URL.Path == "/st2138-api/v1/1/value/status":
			writeRESTProtoResponse(t, w, http.StatusOK, &st2138pb.Value{Kind: &st2138pb.Value_StringValue{StringValue: "ready"}})
		case r.Method == http.MethodPut && r.URL.Path == "/st2138-api/v1/1/value/nested/counter":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read SetValue body: %v", err)
			}
			value := &st2138pb.Value{}
			if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(body, value); err != nil {
				t.Errorf("decode SetValue body: %v", err)
			} else if value.GetInt32Value() != 12 {
				t.Errorf("SetValue int32 = %d, want 12", value.GetInt32Value())
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/st2138-api/v1/1/command/nested/reset":
			if got := r.URL.Query().Get("respond"); got != "true" {
				t.Errorf("respond query = %q, want true", got)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read command body: %v", err)
			}
			if len(body) != 0 {
				t.Errorf("command body = %s, want empty body for zero-argument command", body)
			}
			writeRESTProtoResponse(t, w, http.StatusOK, &st2138pb.CommandResponse{
				Kind: &st2138pb.CommandResponse_NoResponse{NoResponse: &st2138pb.Empty{}},
			})
		default:
			http.Error(w, fmt.Sprintf("unexpected request %s %s", r.Method, r.URL.String()), http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := &Client{Endpoint: strings.TrimPrefix(server.URL, "http://"), Transport: "rest"}
	ctx := context.Background()

	snapshot, err := c.GetDeviceSnapshot(ctx, 1)
	if err != nil {
		t.Fatalf("GetDeviceSnapshot() error = %v", err)
	}
	if snapshot.Parameters["counter"] != "2" || snapshot.FullParameters["locked"] != "fixed" || snapshot.Commands["reset"] != "" {
		t.Errorf("unexpected device snapshot: %#v", snapshot)
	}

	descriptor, err := c.GetParamDescriptor(ctx, 1, "/nested/counter")
	if err != nil {
		t.Fatalf("GetParamDescriptor() error = %v", err)
	}
	if descriptor.GetType() != st2138pb.ParamType_INT32 {
		t.Errorf("descriptor type = %v, want INT32", descriptor.GetType())
	}

	value, err := c.GetRawValue(ctx, 1, "/nested/counter")
	if err != nil {
		t.Fatalf("GetRawValue() error = %v", err)
	}
	if value.GetInt32Value() != 7 {
		t.Errorf("GetRawValue() = %d, want 7", value.GetInt32Value())
	}
	status, err := c.GetStringValue(ctx, 1, "status")
	if err != nil {
		t.Fatalf("GetStringValue() error = %v", err)
	}
	if status != "ready" {
		t.Errorf("GetStringValue() = %q, want ready", status)
	}

	if err := c.SetRawValue(ctx, 1, "/nested/counter", &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: 12}}); err != nil {
		t.Fatalf("SetRawValue() error = %v", err)
	}
	if err := c.ExecuteCommand(ctx, 1, "/nested/reset", &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}}); err != nil {
		t.Fatalf("ExecuteCommand() error = %v", err)
	}
}

func TestClient_RESTRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "parameter not found", http.StatusNotFound)
	}))
	defer server.Close()

	c := &Client{Endpoint: server.URL, Transport: "rest"}
	_, err := c.GetRawValue(context.Background(), 0, "missing")
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") || !strings.Contains(err.Error(), "parameter not found") {
		t.Fatalf("GetRawValue() error = %v, want HTTP 404 with response body", err)
	}
}

func writeRESTProtoResponse(t *testing.T, w http.ResponseWriter, status int, message proto.Message) {
	t.Helper()
	body, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(message)
	if err != nil {
		t.Fatalf("marshal REST response: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
