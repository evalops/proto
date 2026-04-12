package proto_test

import (
	"encoding/json"
	"testing"
	"time"

	eventsv1 "github.com/evalops/proto/gen/go/events/v1"
	memoryv1 "github.com/evalops/proto/gen/go/memory/v1"
	tapv1 "github.com/evalops/proto/gen/go/tap/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMemoryStoreRequestProtoJSONUsesStableProtoFieldNames(t *testing.T) {
	t.Parallel()

	payload := &memoryv1.StoreRequest{
		Scope:      memoryv1.Scope_SCOPE_PROJECT,
		Content:    "Keep PRs focused.",
		Type:       "project",
		Source:     "maestro",
		Confidence: 0.91,
		ProjectId:  "maestro",
		TeamId:     "team-platform",
		Repository: "evalops/maestro",
		Agent:      "maestro",
		IsPolicy:   true,
	}

	encoded, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal StoreRequest: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode StoreRequest JSON: %v", err)
	}

	for _, field := range []string{
		"scope",
		"content",
		"type",
		"source",
		"confidence",
		"project_id",
		"team_id",
		"repository",
		"agent",
		"is_policy",
	} {
		if _, ok := decoded[field]; !ok {
			t.Fatalf("expected proto JSON to contain %q, got %s", field, string(encoded))
		}
	}

	for _, field := range []string{"projectId", "teamId", "isPolicy"} {
		if _, ok := decoded[field]; ok {
			t.Fatalf("expected proto JSON to omit camelCase field %q, got %s", field, string(encoded))
		}
	}
}

func TestMemoryRecallResponseRoundTripPreservesNestedMetadata(t *testing.T) {
	t.Parallel()

	now := timestamppb.New(time.Date(2026, time.April, 12, 15, 0, 0, 0, time.UTC))
	message := &memoryv1.RecallResponse{
		Results: []*memoryv1.RecallResult{
			{
				Memory: &memoryv1.Memory{
					Id:         "mem_123",
					Scope:      memoryv1.Scope_SCOPE_PROJECT,
					Content:    "Retry with exponential backoff.",
					Type:       "project",
					Source:     "maestro",
					Confidence: 0.87,
					ProjectId:  "maestro",
					TeamId:     "team-platform",
					Repository: "evalops/maestro",
					Agent:      "maestro",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
				Similarity: 0.82,
			},
		},
	}

	wire, err := proto.Marshal(message)
	if err != nil {
		t.Fatalf("marshal RecallResponse: %v", err)
	}

	var decoded memoryv1.RecallResponse
	if err := proto.Unmarshal(wire, &decoded); err != nil {
		t.Fatalf("unmarshal RecallResponse: %v", err)
	}

	if len(decoded.Results) != 1 {
		t.Fatalf("expected 1 recall result, got %d", len(decoded.Results))
	}

	result := decoded.Results[0]
	if result.GetMemory().GetProjectId() != "maestro" {
		t.Fatalf("expected project_id to round-trip, got %q", result.GetMemory().GetProjectId())
	}
	if result.GetMemory().GetRepository() != "evalops/maestro" {
		t.Fatalf("expected repository to round-trip, got %q", result.GetMemory().GetRepository())
	}
	if result.GetSimilarity() != float32(0.82) {
		t.Fatalf("expected similarity 0.82, got %v", result.GetSimilarity())
	}
}

func TestCloudEventRoundTripPreservesTypedChangePayload(t *testing.T) {
	t.Parallel()

	recordedAt := timestamppb.New(time.Date(2026, time.April, 12, 15, 10, 0, 0, time.UTC))
	changePayload := &eventsv1.Change{
		Seq:              42,
		OrganizationId:   "org_123",
		AggregateType:    "conversation",
		AggregateId:      "conv_456",
		Operation:        "updated",
		ActorType:        "agent",
		ActorId:          "maestro",
		AggregateVersion: 7,
		RecordedAt:       recordedAt,
		Payload: mustStruct(t, map[string]any{
			"branch_count": 2,
			"visibility":   "workspace",
		}),
	}

	anyPayload, err := anypb.New(changePayload)
	if err != nil {
		t.Fatalf("pack Change payload: %v", err)
	}

	envelope := &eventsv1.CloudEvent{
		SpecVersion:     "1.0",
		Id:              "evt_123",
		Type:            "conversation.updated",
		Source:          "maestro",
		Subject:         "conversation/conv_456",
		Time:            recordedAt,
		DataContentType: "application/protobuf",
		TenantId:        "org_123",
		Data:            anyPayload,
		Extensions: map[string]*structpb.Value{
			"dataschema": structpb.NewStringValue("buf.build/evalops/proto/events.v1.Change"),
		},
	}

	wire, err := proto.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal CloudEvent: %v", err)
	}

	var decoded eventsv1.CloudEvent
	if err := proto.Unmarshal(wire, &decoded); err != nil {
		t.Fatalf("unmarshal CloudEvent: %v", err)
	}

	if decoded.GetData().GetTypeUrl() != "type.googleapis.com/events.v1.Change" {
		t.Fatalf("unexpected Change type URL %q", decoded.GetData().GetTypeUrl())
	}

	var unpacked eventsv1.Change
	if err := decoded.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack Change payload: %v", err)
	}

	if unpacked.GetAggregateId() != "conv_456" {
		t.Fatalf("expected aggregate_id conv_456, got %q", unpacked.GetAggregateId())
	}
	if unpacked.GetPayload().GetFields()["visibility"].GetStringValue() != "workspace" {
		t.Fatalf("expected payload visibility workspace, got %#v", unpacked.GetPayload().GetFields()["visibility"])
	}
}

func TestCloudEventRoundTripPreservesTypedTapPayload(t *testing.T) {
	t.Parallel()

	providerTimestamp := timestamppb.New(time.Date(2026, time.April, 12, 15, 20, 0, 0, time.UTC))
	tapPayload := &tapv1.TapEventData{
		Provider:   "hubspot",
		EntityType: "deal",
		EntityId:   "deal_123",
		Action:     "updated",
		RequestId:  "req_789",
		TenantId:   "org_123",
		Snapshot:   mustStruct(t, map[string]any{"stage": "qualified"}),
		Changes: map[string]*tapv1.FieldChange{
			"stage": {
				From: structpb.NewStringValue("new"),
				To:   structpb.NewStringValue("qualified"),
			},
		},
		ProviderEventId:   "evt_provider_1",
		ProviderTimestamp: providerTimestamp,
	}

	anyPayload, err := anypb.New(tapPayload)
	if err != nil {
		t.Fatalf("pack TapEventData payload: %v", err)
	}

	envelope := &eventsv1.CloudEvent{
		SpecVersion:     "1.0",
		Id:              "evt_456",
		Type:            "tap.entity.updated",
		Source:          "ensemble-tap",
		Subject:         "deal/deal_123",
		Time:            providerTimestamp,
		DataContentType: "application/protobuf",
		TenantId:        "org_123",
		Data:            anyPayload,
	}

	wire, err := proto.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal CloudEvent: %v", err)
	}

	var decoded eventsv1.CloudEvent
	if err := proto.Unmarshal(wire, &decoded); err != nil {
		t.Fatalf("unmarshal CloudEvent: %v", err)
	}

	if decoded.GetData().GetTypeUrl() != "type.googleapis.com/tap.v1.TapEventData" {
		t.Fatalf("unexpected TapEventData type URL %q", decoded.GetData().GetTypeUrl())
	}

	var unpacked tapv1.TapEventData
	if err := decoded.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack TapEventData payload: %v", err)
	}

	if unpacked.GetProvider() != "hubspot" {
		t.Fatalf("expected provider hubspot, got %q", unpacked.GetProvider())
	}
	if unpacked.GetChanges()["stage"].GetTo().GetStringValue() != "qualified" {
		t.Fatalf("expected stage transition to qualified, got %#v", unpacked.GetChanges()["stage"].GetTo())
	}
}

func mustStruct(t *testing.T, value map[string]any) *structpb.Struct {
	t.Helper()

	message, err := structpb.NewStruct(value)
	if err != nil {
		t.Fatalf("build struct payload: %v", err)
	}
	return message
}
