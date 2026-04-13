package proto_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	approvalsv1 "github.com/evalops/proto/gen/go/approvals/v1"
	configv1 "github.com/evalops/proto/gen/go/config/v1"
	connectorsv1 "github.com/evalops/proto/gen/go/connectors/v1"
	entitiesv1 "github.com/evalops/proto/gen/go/entities/v1"
	eventsv1 "github.com/evalops/proto/gen/go/events/v1"
	governancev1 "github.com/evalops/proto/gen/go/governance/v1"
	keysv1 "github.com/evalops/proto/gen/go/keys/v1"
	memoryv1 "github.com/evalops/proto/gen/go/memory/v1"
	meterv1 "github.com/evalops/proto/gen/go/meter/v1"
	notificationsv1 "github.com/evalops/proto/gen/go/notifications/v1"
	objectivesv1 "github.com/evalops/proto/gen/go/objectives/v1"
	skillsv1 "github.com/evalops/proto/gen/go/skills/v1"
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

func TestFeatureFlagSnapshotProtoJSONUsesStableProtoFieldNames(t *testing.T) {
	t.Parallel()

	payload := &configv1.FeatureFlagSnapshot{
		SchemaVersion: 1,
		Flags: []*configv1.FeatureFlag{
			{
				Key:            "platform.kill_switches.llm_gateway.inference",
				Enabled:        true,
				RolloutPercent: 100,
				Owners:         []string{"platform-app"},
				Description:    "Master kill switch for managed llm-gateway inference execution.",
			},
		},
	}

	encoded, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal FeatureFlagSnapshot: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode FeatureFlagSnapshot JSON: %v", err)
	}

	for _, field := range []string{"schema_version", "flags"} {
		if _, ok := decoded[field]; !ok {
			t.Fatalf("expected proto JSON to contain %q, got %s", field, string(encoded))
		}
	}
	if _, ok := decoded["schemaVersion"]; ok {
		t.Fatalf("expected proto JSON to omit camelCase schemaVersion, got %s", string(encoded))
	}
}

func TestFeatureFlagSnapshotFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message configv1.FeatureFlagSnapshot
	loadProtoJSONFixture(t, filepath.Join("proto", "config", "v1", "testdata", "feature_flag_snapshot.json"), &message)

	if message.GetSchemaVersion() != 1 {
		t.Fatalf("expected schema_version 1, got %d", message.GetSchemaVersion())
	}
	if len(message.GetFlags()) != 6 {
		t.Fatalf("expected 6 flags, got %d", len(message.GetFlags()))
	}
	if message.GetFlags()[0].GetKey() != "llm_gateway.model_routing.provider_failover" {
		t.Fatalf("unexpected first flag key %q", message.GetFlags()[0].GetKey())
	}
	if !message.GetFlags()[1].GetEnabled() {
		t.Fatal("expected second flag to be enabled")
	}
	if message.GetFlags()[2].GetEnabled() {
		t.Fatal("expected third flag to be disabled")
	}
	if message.GetFlags()[5].GetRolloutPercent() != 100 {
		t.Fatalf("expected rollout_percent 100, got %d", message.GetFlags()[5].GetRolloutPercent())
	}
}

func TestMemoryStoreRequestFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message memoryv1.StoreRequest
	loadProtoJSONFixture(t, filepath.Join("proto", "memory", "v1", "testdata", "store_request.json"), &message)

	if message.GetScope() != memoryv1.Scope_SCOPE_PROJECT {
		t.Fatalf("expected SCOPE_PROJECT, got %v", message.GetScope())
	}
	if message.GetProjectId() != "maestro" {
		t.Fatalf("expected project_id maestro, got %q", message.GetProjectId())
	}
	if !message.GetIsPolicy() {
		t.Fatal("expected is_policy=true in store fixture")
	}
}

func TestMemoryRecallResponseFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message memoryv1.RecallResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "memory", "v1", "testdata", "recall_response.json"), &message)

	if len(message.GetResults()) != 1 {
		t.Fatalf("expected 1 recall result, got %d", len(message.GetResults()))
	}

	result := message.GetResults()[0]
	if result.GetMemory().GetRepository() != "evalops/maestro" {
		t.Fatalf("expected repository evalops/maestro, got %q", result.GetMemory().GetRepository())
	}
	if result.GetSimilarity() != float32(0.82) {
		t.Fatalf("expected similarity 0.82, got %v", result.GetSimilarity())
	}
}

func TestMeterRecordUsageRequestFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message meterv1.RecordUsageRequest
	loadProtoJSONFixture(t, filepath.Join("proto", "meter", "v1", "testdata", "record_usage_request.json"), &message)

	if message.GetTeamId() != "team_eng" {
		t.Fatalf("expected team_id team_eng, got %q", message.GetTeamId())
	}
	if message.GetEventType() != "llm.completion" {
		t.Fatalf("expected event_type llm.completion, got %q", message.GetEventType())
	}
	if message.GetMetadata().GetFields()["pipeline_deal_id"].GetStringValue() != "deal_123" {
		t.Fatalf("expected pipeline_deal_id deal_123, got %#v", message.GetMetadata().GetFields()["pipeline_deal_id"])
	}
	if message.GetData().GetFields()["temperature"].GetNumberValue() != 0.2 {
		t.Fatalf("expected temperature 0.2, got %#v", message.GetData().GetFields()["temperature"])
	}
}

func TestMeterRecordUsageResponseFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message meterv1.RecordUsageResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "meter", "v1", "testdata", "record_usage_response.json"), &message)

	record := message.GetRecord()
	if record.GetOrganizationId() != "org_123" {
		t.Fatalf("expected organization_id org_123, got %q", record.GetOrganizationId())
	}
	if record.GetCreatedAt().AsTime().Format(time.RFC3339) != "2026-04-13T08:15:01Z" {
		t.Fatalf("unexpected created_at %s", record.GetCreatedAt().AsTime().Format(time.RFC3339))
	}
	if record.GetMetadata().GetFields()["pipeline_deal_id"].GetStringValue() != "deal_123" {
		t.Fatalf("expected pipeline_deal_id deal_123, got %#v", record.GetMetadata().GetFields()["pipeline_deal_id"])
	}
}

func TestMeterQueryUsageResponseFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message meterv1.UsageQueryResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "meter", "v1", "testdata", "query_usage_response.json"), &message)

	if len(message.GetRecords()) != 2 {
		t.Fatalf("expected 2 records, got %d", len(message.GetRecords()))
	}
	if message.GetRecords()[1].GetProvider() != "openai" {
		t.Fatalf("expected second provider openai, got %q", message.GetRecords()[1].GetProvider())
	}
	if message.GetRecords()[1].GetData().GetFields()["temperature"].GetNumberValue() != 0.7 {
		t.Fatalf("expected second record temperature 0.7, got %#v", message.GetRecords()[1].GetData().GetFields()["temperature"])
	}
}

func TestMeterUsageSummaryResponseFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message meterv1.UsageSummaryResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "meter", "v1", "testdata", "usage_summary_response.json"), &message)

	if len(message.GetBuckets()) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(message.GetBuckets()))
	}
	if message.GetBuckets()[0].GetKey() != "claude-opus-4.6" {
		t.Fatalf("unexpected first bucket key %q", message.GetBuckets()[0].GetKey())
	}
	if message.GetBuckets()[0].GetTotalCostUsd() != 0.0142 {
		t.Fatalf("unexpected first bucket cost %v", message.GetBuckets()[0].GetTotalCostUsd())
	}
}

func TestMeterSummaryResponseFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message meterv1.MeterSummaryResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "meter", "v1", "testdata", "meter_summary_response.json"), &message)

	if message.GetMeterId() != "input_tokens" {
		t.Fatalf("expected meter_id input_tokens, got %q", message.GetMeterId())
	}
	if len(message.GetBuckets()) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(message.GetBuckets()))
	}
	if message.GetBuckets()[0].GetValue() != 125 {
		t.Fatalf("expected first bucket value 125, got %v", message.GetBuckets()[0].GetValue())
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

func TestCloudEventChangeFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message eventsv1.CloudEvent
	loadProtoJSONFixture(t, filepath.Join("proto", "events", "v1", "testdata", "cloud_event_change.json"), &message)

	if message.GetSubject() != "conversation/conv_456" {
		t.Fatalf("expected subject conversation/conv_456, got %q", message.GetSubject())
	}
	if message.GetTenantId() != "org_123" {
		t.Fatalf("expected tenant_id org_123, got %q", message.GetTenantId())
	}
	if message.GetDataContentType() != "application/protobuf" {
		t.Fatalf("expected data_content_type application/protobuf, got %q", message.GetDataContentType())
	}

	if message.GetData().GetTypeUrl() != "type.googleapis.com/events.v1.Change" {
		t.Fatalf("unexpected Change type URL %q", message.GetData().GetTypeUrl())
	}

	var unpacked eventsv1.Change
	if err := message.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack Change payload: %v", err)
	}

	if unpacked.GetAggregateId() != "conv_456" {
		t.Fatalf("expected aggregate_id conv_456, got %q", unpacked.GetAggregateId())
	}
	if unpacked.GetPayload().GetFields()["branch_count"].GetNumberValue() != 2 {
		t.Fatalf("expected branch_count 2, got %#v", unpacked.GetPayload().GetFields()["branch_count"])
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

func TestCloudEventTapFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message eventsv1.CloudEvent
	loadProtoJSONFixture(t, filepath.Join("proto", "events", "v1", "testdata", "cloud_event_tap.json"), &message)

	if message.GetSubject() != "deal/deal_123" {
		t.Fatalf("expected subject deal/deal_123, got %q", message.GetSubject())
	}
	if message.GetTenantId() != "org_123" {
		t.Fatalf("expected tenant_id org_123, got %q", message.GetTenantId())
	}
	if message.GetDataContentType() != "application/protobuf" {
		t.Fatalf("expected data_content_type application/protobuf, got %q", message.GetDataContentType())
	}

	if message.GetData().GetTypeUrl() != "type.googleapis.com/tap.v1.TapEventData" {
		t.Fatalf("unexpected TapEventData type URL %q", message.GetData().GetTypeUrl())
	}

	var unpacked tapv1.TapEventData
	if err := message.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack TapEventData payload: %v", err)
	}

	if unpacked.GetProvider() != "hubspot" {
		t.Fatalf("expected provider hubspot, got %q", unpacked.GetProvider())
	}
	if unpacked.GetRequestId() != "req_789" {
		t.Fatalf("expected request_id req_789, got %q", unpacked.GetRequestId())
	}
	if unpacked.GetChanges()["stage"].GetTo().GetStringValue() != "qualified" {
		t.Fatalf("expected stage transition to qualified, got %#v", unpacked.GetChanges()["stage"].GetTo())
	}
}

func TestCloudEventTapHubspotDealQualifiedFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message eventsv1.CloudEvent
	loadProtoJSONFixture(t, filepath.Join("proto", "events", "v1", "testdata", "cloud_event_tap_hubspot_deal_qualified.json"), &message)

	if message.GetType() != "ensemble.tap.hubspot.deal.updated" {
		t.Fatalf("expected type ensemble.tap.hubspot.deal.updated, got %q", message.GetType())
	}
	if message.GetSubject() != "hubspot/deal/deal_123" {
		t.Fatalf("expected subject hubspot/deal/deal_123, got %q", message.GetSubject())
	}
	if message.GetTenantId() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected tenant_id 11111111-1111-1111-1111-111111111111, got %q", message.GetTenantId())
	}

	var unpacked tapv1.TapEventData
	if err := message.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack TapEventData payload: %v", err)
	}

	if unpacked.GetProvider() != "hubspot" {
		t.Fatalf("expected provider hubspot, got %q", unpacked.GetProvider())
	}
	if unpacked.GetTenantId() != message.GetTenantId() {
		t.Fatalf("expected payload tenant_id %q, got %q", message.GetTenantId(), unpacked.GetTenantId())
	}
	if unpacked.GetSnapshot().AsMap()["company_domain"] != "acme.com" {
		t.Fatalf("expected company_domain acme.com, got %#v", unpacked.GetSnapshot().AsMap()["company_domain"])
	}
	if unpacked.GetChanges()["stage"].GetTo().GetStringValue() != "qualified" {
		t.Fatalf("expected stage transition to qualified, got %#v", unpacked.GetChanges()["stage"].GetTo())
	}
}

func TestCloudEventPipelineActivityCreateRepliedFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message eventsv1.CloudEvent
	loadProtoJSONFixture(t, filepath.Join("proto", "events", "v1", "testdata", "cloud_event_pipeline_activity_create_replied.json"), &message)

	if message.GetSubject() != "pipeline.changes.activity.create" {
		t.Fatalf("expected subject pipeline.changes.activity.create, got %q", message.GetSubject())
	}
	if message.GetTenantId() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected tenant_id 11111111-1111-1111-1111-111111111111, got %q", message.GetTenantId())
	}
	if message.GetDataContentType() != "application/protobuf" {
		t.Fatalf("expected data_content_type application/protobuf, got %q", message.GetDataContentType())
	}

	var unpacked eventsv1.Change
	if err := message.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack Change payload: %v", err)
	}

	if unpacked.GetAggregateType() != "activity" || unpacked.GetOperation() != "create" {
		t.Fatalf("expected activity/create, got %q/%q", unpacked.GetAggregateType(), unpacked.GetOperation())
	}

	payload := unpacked.GetPayload().AsMap()
	if payload["owner_type"] != "contact" {
		t.Fatalf("expected owner_type contact, got %#v", payload["owner_type"])
	}
	if payload["outcome"] != "replied" {
		t.Fatalf("expected outcome replied, got %#v", payload["outcome"])
	}
	if payload["channel"] != "email" {
		t.Fatalf("expected channel email, got %#v", payload["channel"])
	}
}

func TestCloudEventPipelineSignalCreateLinkedInActiveFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message eventsv1.CloudEvent
	loadProtoJSONFixture(t, filepath.Join("proto", "events", "v1", "testdata", "cloud_event_pipeline_signal_create_linkedin_active.json"), &message)

	if message.GetSubject() != "pipeline.changes.signal.create" {
		t.Fatalf("expected subject pipeline.changes.signal.create, got %q", message.GetSubject())
	}
	if message.GetTenantId() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected tenant_id 11111111-1111-1111-1111-111111111111, got %q", message.GetTenantId())
	}
	if message.GetDataContentType() != "application/protobuf" {
		t.Fatalf("expected data_content_type application/protobuf, got %q", message.GetDataContentType())
	}

	var unpacked eventsv1.Change
	if err := message.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack Change payload: %v", err)
	}

	if unpacked.GetAggregateType() != "signal" || unpacked.GetOperation() != "create" {
		t.Fatalf("expected signal/create, got %q/%q", unpacked.GetAggregateType(), unpacked.GetOperation())
	}

	payload := unpacked.GetPayload().AsMap()
	if payload["signal_type"] != "linkedin_active" {
		t.Fatalf("expected signal_type linkedin_active, got %#v", payload["signal_type"])
	}
	if payload["source"] != "linkedin" {
		t.Fatalf("expected source linkedin, got %#v", payload["source"])
	}
	if payload["strength"] != float64(87) {
		t.Fatalf("expected strength 87, got %#v", payload["strength"])
	}
}

func TestCloudEventPipelineDealUpdateClosedWonFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message eventsv1.CloudEvent
	loadProtoJSONFixture(t, filepath.Join("proto", "events", "v1", "testdata", "cloud_event_pipeline_deal_update_closed_won.json"), &message)

	if message.GetSubject() != "pipeline.changes.deal.update" {
		t.Fatalf("expected subject pipeline.changes.deal.update, got %q", message.GetSubject())
	}
	if message.GetTenantId() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected tenant_id 11111111-1111-1111-1111-111111111111, got %q", message.GetTenantId())
	}
	if message.GetDataContentType() != "application/protobuf" {
		t.Fatalf("expected data_content_type application/protobuf, got %q", message.GetDataContentType())
	}

	var unpacked eventsv1.Change
	if err := message.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack Change payload: %v", err)
	}

	if unpacked.GetAggregateType() != "deal" || unpacked.GetOperation() != "update" {
		t.Fatalf("expected deal/update, got %q/%q", unpacked.GetAggregateType(), unpacked.GetOperation())
	}

	payload := unpacked.GetPayload().AsMap()
	if payload["stage"] != "closed_won" {
		t.Fatalf("expected stage closed_won, got %#v", payload["stage"])
	}
	if payload["title"] != "Acme expansion" {
		t.Fatalf("expected title Acme expansion, got %#v", payload["title"])
	}
	if payload["value"] != float64(120000) {
		t.Fatalf("expected value 120000, got %#v", payload["value"])
	}
}

func TestCloudEventEvaluationCompletedTechnicalCapabilityFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message eventsv1.CloudEvent
	loadProtoJSONFixture(t, filepath.Join("proto", "events", "v1", "testdata", "cloud_event_evaluation_completed_technical_capability.json"), &message)

	if message.GetType() != "evaluation.completed" {
		t.Fatalf("expected type evaluation.completed, got %q", message.GetType())
	}
	if message.GetSubject() != "product.evaluation.completed" {
		t.Fatalf("expected subject product.evaluation.completed, got %q", message.GetSubject())
	}
	if message.GetTenantId() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected tenant_id 11111111-1111-1111-1111-111111111111, got %q", message.GetTenantId())
	}
	if got := message.GetExtensions()["dataschema"].GetStringValue(); got != "buf.build/evalops/proto/events.v1.EvaluationCompleted" {
		t.Fatalf("expected dataschema buf.build/evalops/proto/events.v1.EvaluationCompleted, got %q", got)
	}

	var unpacked eventsv1.EvaluationCompleted
	if err := message.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack EvaluationCompleted payload: %v", err)
	}
	if unpacked.GetSignalType() != "technical_capability" {
		t.Fatalf("expected signal_type technical_capability, got %q", unpacked.GetSignalType())
	}
	if unpacked.GetRun().GetId() != "run-1" {
		t.Fatalf("expected run.id run-1, got %q", unpacked.GetRun().GetId())
	}
	if unpacked.GetMetrics().GetSuccessRate() != 0.9 {
		t.Fatalf("expected metrics.success_rate 0.9, got %v", unpacked.GetMetrics().GetSuccessRate())
	}
	if got := unpacked.GetCompanyDomains(); len(got) != 1 || got[0] != "acme.com" {
		t.Fatalf("expected company_domains [acme.com], got %#v", got)
	}
}

func TestCloudEventParkerWorkRelationshipUpdateTerminatedFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message eventsv1.CloudEvent
	loadProtoJSONFixture(t, filepath.Join("proto", "events", "v1", "testdata", "cloud_event_parker_work_relationship_update_terminated.json"), &message)

	if message.GetSubject() != "parker.changes.work_relationship.update" {
		t.Fatalf("expected subject parker.changes.work_relationship.update, got %q", message.GetSubject())
	}
	if message.GetTenantId() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected tenant_id 11111111-1111-1111-1111-111111111111, got %q", message.GetTenantId())
	}

	var unpacked eventsv1.Change
	if err := message.GetData().UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("unpack Change payload: %v", err)
	}

	if unpacked.GetAggregateType() != "work_relationship" || unpacked.GetOperation() != "update" {
		t.Fatalf("expected work_relationship/update, got %q/%q", unpacked.GetAggregateType(), unpacked.GetOperation())
	}

	payload := unpacked.GetPayload().AsMap()
	if payload["status"] != "terminated" {
		t.Fatalf("expected status terminated, got %#v", payload["status"])
	}
	if payload["termination_reason"] != "voluntary" {
		t.Fatalf("expected termination_reason voluntary, got %#v", payload["termination_reason"])
	}
	if payload["employment_type"] != "full_time" {
		t.Fatalf("expected employment_type full_time, got %#v", payload["employment_type"])
	}
}

func TestApprovalsRequestApprovalFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message approvalsv1.RequestApprovalRequest
	loadProtoJSONFixture(t, filepath.Join("proto", "approvals", "v1", "testdata", "request_approval_request.json"), &message)

	if message.GetWorkspaceId() != "ws_approval" {
		t.Fatalf("expected workspace_id ws_approval, got %q", message.GetWorkspaceId())
	}
	if message.GetRiskLevel() != approvalsv1.RiskLevel_RISK_LEVEL_HIGH {
		t.Fatalf("expected RISK_LEVEL_HIGH, got %v", message.GetRiskLevel())
	}
	if string(message.GetActionPayload()) != `{"path":"/tmp/secret.txt","recursive":false}` {
		t.Fatalf("unexpected action_payload %q", string(message.GetActionPayload()))
	}
}

func TestConnectorsRegisterConnectionFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message connectorsv1.RegisterConnectionRequest
	loadProtoJSONFixture(t, filepath.Join("proto", "connectors", "v1", "testdata", "register_connection_request.json"), &message)

	if message.GetProviderId() != "hubspot" {
		t.Fatalf("expected provider_id hubspot, got %q", message.GetProviderId())
	}
	if message.GetAuthType() != connectorsv1.AuthType_AUTH_TYPE_OAUTH2 {
		t.Fatalf("expected AUTH_TYPE_OAUTH2, got %v", message.GetAuthType())
	}
	if message.GetCredentials()["secret_ref"] != "gsm://evalops/hubspot-client-secret" {
		t.Fatalf("expected secret_ref credential, got %#v", message.GetCredentials())
	}
}

func TestEntitiesGetCanonicalFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message entitiesv1.GetCanonicalResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "entities", "v1", "testdata", "get_canonical_response.json"), &message)

	entity := message.GetEntity()
	if entity.GetPrimaryType() != entitiesv1.EntityType_ENTITY_TYPE_CONTACT {
		t.Fatalf("expected ENTITY_TYPE_CONTACT, got %v", entity.GetPrimaryType())
	}
	if len(entity.GetRefs()) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(entity.GetRefs()))
	}
	if entity.GetRefs()[0].GetEmails()[0] != "jamie@example.com" {
		t.Fatalf("expected primary email jamie@example.com, got %#v", entity.GetRefs()[0].GetEmails())
	}
}

func TestGovernanceEvaluateActionFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message governancev1.EvaluateActionRequest
	loadProtoJSONFixture(t, filepath.Join("proto", "governance", "v1", "testdata", "evaluate_action_request.json"), &message)

	if message.GetWorkspaceId() != "ws_governance" {
		t.Fatalf("expected workspace_id ws_governance, got %q", message.GetWorkspaceId())
	}
	if string(message.GetActionPayload()) != `{"credential":"sk-live-123","target":"slack"}` {
		t.Fatalf("unexpected action_payload %q", string(message.GetActionPayload()))
	}
}

func TestKeysResolveProviderRefFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var request keysv1.ResolveProviderRefRequest
	loadProtoJSONFixture(t, filepath.Join("proto", "keys", "v1", "testdata", "resolve_provider_ref_request.json"), &request)
	if request.GetProvider() != "openai" {
		t.Fatalf("expected provider openai, got %q", request.GetProvider())
	}
	if request.GetEnvironment() != "production" {
		t.Fatalf("expected environment production, got %q", request.GetEnvironment())
	}
	if request.GetTeamId() != "team_platform" {
		t.Fatalf("expected team_id team_platform, got %q", request.GetTeamId())
	}

	var response keysv1.ResolveProviderRefResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "keys", "v1", "testdata", "resolve_provider_ref_response.json"), &response)
	if response.GetProviderRef().GetEndpointUrl() != "https://api.openai.com/v1" {
		t.Fatalf("expected endpoint_url https://api.openai.com/v1, got %q", response.GetProviderRef().GetEndpointUrl())
	}
	if response.GetProviderRef().GetCredentialData().GetFields()["api_key"].GetStringValue() != "sk-live-123" {
		t.Fatalf("unexpected credential_data api_key %#v", response.GetProviderRef().GetCredentialData().GetFields()["api_key"])
	}
}

func TestNotificationsGetPreferencesFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message notificationsv1.GetPreferencesResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "notifications", "v1", "testdata", "get_preferences_response.json"), &message)

	preferences := message.GetPreferences()
	if preferences.GetDefaultChannel() != notificationsv1.DeliveryChannel_DELIVERY_CHANNEL_SLACK {
		t.Fatalf("expected DELIVERY_CHANNEL_SLACK, got %v", preferences.GetDefaultChannel())
	}
	if len(preferences.GetEscalationRules()) != 2 {
		t.Fatalf("expected 2 escalation rules, got %d", len(preferences.GetEscalationRules()))
	}
	if preferences.GetEscalationRules()[1].GetEscalateToChannel() != notificationsv1.DeliveryChannel_DELIVERY_CHANNEL_EMAIL {
		t.Fatalf("expected second escalation rule to target email, got %v", preferences.GetEscalationRules()[1].GetEscalateToChannel())
	}
}

func TestObjectivesCreateResponseFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message objectivesv1.CreateResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "objectives", "v1", "testdata", "create_response.json"), &message)

	objective := message.GetObjective()
	if objective.GetState() != objectivesv1.ObjectiveState_OBJECTIVE_STATE_RUNNING {
		t.Fatalf("expected OBJECTIVE_STATE_RUNNING, got %v", objective.GetState())
	}
	if len(objective.GetProvenance()) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(objective.GetProvenance()))
	}
	if len(objective.GetMutations()) != 1 {
		t.Fatalf("expected 1 mutation record, got %d", len(objective.GetMutations()))
	}
	if objective.GetMutations()[0].GetStatus() != objectivesv1.MutationStatus_MUTATION_STATUS_APPROVED {
		t.Fatalf("expected first mutation to be approved, got %v", objective.GetMutations()[0].GetStatus())
	}
}

func TestSkillsSearchResponseFixtureMatchesProtoContract(t *testing.T) {
	t.Parallel()

	var message skillsv1.SearchResponse
	loadProtoJSONFixture(t, filepath.Join("proto", "skills", "v1", "testdata", "search_response.json"), &message)

	if message.GetTotal() != 2 {
		t.Fatalf("expected total 2, got %d", message.GetTotal())
	}
	if len(message.GetSkills()) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(message.GetSkills()))
	}
	if message.GetSkills()[0].GetScope() != skillsv1.SkillScope_SKILL_SCOPE_WORKSPACE {
		t.Fatalf("expected first skill to be workspace-scoped, got %v", message.GetSkills()[0].GetScope())
	}
	if len(message.GetSkills()[1].GetTags()) != 2 {
		t.Fatalf("expected second skill to have 2 tags, got %#v", message.GetSkills()[1].GetTags())
	}
	if message.GetSkills()[1].GetTags()[1] != "metrics" {
		t.Fatalf("expected second skill second tag to be metrics, got %#v", message.GetSkills()[1].GetTags())
	}
}

func loadProtoJSONFixture(t *testing.T, path string, message proto.Message) {
	t.Helper()

	fixture, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(fixture, message); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", path, err)
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
