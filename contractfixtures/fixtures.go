package contractfixtures

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/evalops/proto/eventhelpers"
	configv1 "github.com/evalops/proto/gen/go/config/v1"
	eventsv1 "github.com/evalops/proto/gen/go/events/v1"
	tapv1 "github.com/evalops/proto/gen/go/tap/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	AgentsRegisterResponse                      = "agents/v1/testdata/register_response.json"
	AgentsPushConfigResponse                    = "agents/v1/testdata/push_config_response.json"
	AgentsDelegateResponse                      = "agents/v1/testdata/delegate_response.json"
	ApprovalsRequestApprovalRequest             = "approvals/v1/testdata/request_approval_request.json"
	ConfigFeatureFlagSnapshot                   = "config/v1/testdata/feature_flag_snapshot.json"
	ConnectorsRegisterConnectionRequest         = "connectors/v1/testdata/register_connection_request.json"
	EntitiesGetCanonicalResponse                = "entities/v1/testdata/get_canonical_response.json"
	EventChange                                 = "events/v1/testdata/cloud_event_change.json"
	EventEvaluationCompletedTechnicalCapability = "events/v1/testdata/cloud_event_evaluation_completed_technical_capability.json"
	EventPipelineActivityCreateReplied          = "events/v1/testdata/cloud_event_pipeline_activity_create_replied.json"
	EventPipelineDealUpdateClosedWon            = "events/v1/testdata/cloud_event_pipeline_deal_update_closed_won.json"
	EventPipelineSignalCreateLinkedInActive     = "events/v1/testdata/cloud_event_pipeline_signal_create_linkedin_active.json"
	EventParkerWorkRelationshipUpdateTerminated = "events/v1/testdata/cloud_event_parker_work_relationship_update_terminated.json"
	EventTap                                    = "events/v1/testdata/cloud_event_tap.json"
	EventTapHubspotDealQualified                = "events/v1/testdata/cloud_event_tap_hubspot_deal_qualified.json"
	GovernanceEvaluateActionRequest             = "governance/v1/testdata/evaluate_action_request.json"
	KeysResolveProviderRefRequest               = "keys/v1/testdata/resolve_provider_ref_request.json"
	KeysResolveProviderRefResponse              = "keys/v1/testdata/resolve_provider_ref_response.json"
	MemoryRecallResponse                        = "memory/v1/testdata/recall_response.json"
	MemoryStoreRequest                          = "memory/v1/testdata/store_request.json"
	MeterRecordUsageRequestLLMGatewayResponses  = "meter/v1/testdata/record_usage_request_llm_gateway_responses.json"
	MeterRecordUsageRequest                     = "meter/v1/testdata/record_usage_request.json"
	MeterRecordUsageResponse                    = "meter/v1/testdata/record_usage_response.json"
	MeterUsageQueryResponse                     = "meter/v1/testdata/query_usage_response.json"
	MeterUsageSummaryResponse                   = "meter/v1/testdata/usage_summary_response.json"
	MeterMeterSummaryResponse                   = "meter/v1/testdata/meter_summary_response.json"
	NotificationsGetPreferencesResponse         = "notifications/v1/testdata/get_preferences_response.json"
	ObjectivesCreateResponse                    = "objectives/v1/testdata/create_response.json"
	SkillsSearchResponse                        = "skills/v1/testdata/search_response.json"
	WorkflowsHandleTriggerRequest               = "workflows/v1/testdata/handle_trigger_request.json"
	WorkflowsHandleTriggerResponse              = "workflows/v1/testdata/handle_trigger_response.json"
	WorkflowsPublishVersionResponse             = "workflows/v1/testdata/publish_version_response.json"
	WorkflowsGetRunResponse                     = "workflows/v1/testdata/get_run_response.json"
)

var fixtureCatalog = []string{
	AgentsRegisterResponse,
	AgentsPushConfigResponse,
	AgentsDelegateResponse,
	ApprovalsRequestApprovalRequest,
	ConfigFeatureFlagSnapshot,
	ConnectorsRegisterConnectionRequest,
	EntitiesGetCanonicalResponse,
	EventChange,
	EventEvaluationCompletedTechnicalCapability,
	EventPipelineActivityCreateReplied,
	EventPipelineDealUpdateClosedWon,
	EventPipelineSignalCreateLinkedInActive,
	EventParkerWorkRelationshipUpdateTerminated,
	EventTap,
	EventTapHubspotDealQualified,
	GovernanceEvaluateActionRequest,
	KeysResolveProviderRefRequest,
	KeysResolveProviderRefResponse,
	MemoryRecallResponse,
	MemoryStoreRequest,
	MeterMeterSummaryResponse,
	MeterRecordUsageRequest,
	MeterRecordUsageRequestLLMGatewayResponses,
	MeterRecordUsageResponse,
	MeterUsageQueryResponse,
	MeterUsageSummaryResponse,
	NotificationsGetPreferencesResponse,
	ObjectivesCreateResponse,
	SkillsSearchResponse,
	WorkflowsHandleTriggerRequest,
	WorkflowsHandleTriggerResponse,
	WorkflowsPublishVersionResponse,
	WorkflowsGetRunResponse,
}

var embeddedFixtures = map[string][]byte{
	EventEvaluationCompletedTechnicalCapability: []byte(`{
  "spec_version": "1.0",
  "id": "evt_eval_completed_technical_capability_1",
  "type": "evaluation.completed",
  "source": "fermata",
  "subject": "product.evaluation.completed",
  "time": "2026-04-13T12:00:00Z",
  "data_content_type": "application/protobuf",
  "tenant_id": "11111111-1111-1111-1111-111111111111",
  "data": {
    "@type": "type.googleapis.com/events.v1.EvaluationCompleted",
    "signal_type": "technical_capability",
    "summary": "Claude beats GPT-4 on latency by 40% on enterprise workloads.",
    "success_rate": 0.9,
    "company_domains": [
      "acme.com"
    ],
    "company_names": [
      "Acme Corporation"
    ],
    "deal_ids": [
      "deal-123"
    ],
    "run": {
      "id": "run-1",
      "test_suite_id": "suite-1",
      "test_suite_name": "Latency benchmark",
      "name": "Enterprise latency proof",
      "description": "Claude beats GPT-4 on latency by 40% on enterprise workloads.",
      "tags": [
        "pipeline:signal_type=technical_capability",
        "pipeline:company_domain=acme.com",
        "pipeline:company_name=Acme Corporation",
        "pipeline:deal_id=deal-123"
      ],
      "completed_at": "2026-04-13T12:00:00Z"
    },
    "metrics": {
      "total_tests": "20",
      "passed_tests": "18",
      "failed_tests": "2",
      "total_cost": 4.2,
      "duration": 12.5,
      "success_rate": 0.9
    }
  },
  "extensions": {
    "dataschema": "buf.build/evalops/proto/events.v1.EvaluationCompleted"
  }
}`),
	EventPipelineDealUpdateClosedWon: []byte(`{
  "spec_version": "1.0",
  "id": "evt_pipeline_deal_closed_won_1",
  "type": "pipeline.changes.deal.update",
  "source": "pipeline",
  "subject": "pipeline.changes.deal.update",
  "time": "2026-04-13T09:15:00.345678901Z",
  "data_content_type": "application/protobuf",
  "tenant_id": "11111111-1111-1111-1111-111111111111",
  "data": {
    "@type": "type.googleapis.com/events.v1.Change",
    "seq": "312",
    "organization_id": "11111111-1111-1111-1111-111111111111",
    "aggregate_type": "deal",
    "aggregate_id": "66666666-6666-6666-6666-666666666666",
    "operation": "update",
    "actor_type": "service",
    "actor_id": "pipeline-api",
    "aggregate_version": "9",
    "recorded_at": "2026-04-13T09:15:00.345678901Z",
    "payload": {
      "id": "66666666-6666-6666-6666-666666666666",
      "organization_id": "11111111-1111-1111-1111-111111111111",
      "contact_id": "44444444-4444-4444-4444-444444444444",
      "company_id": "77777777-7777-7777-7777-777777777777",
      "title": "Acme expansion",
      "stage": "closed_won",
      "value": 120000,
      "currency": "USD",
      "probability": 100,
      "expected_close": "2026-04-30",
      "owner_id": "owner_123",
      "version": 9,
      "created_at": "2026-04-10T16:00:00Z",
      "updated_at": "2026-04-13T09:15:00.345678901Z"
    }
  },
  "extensions": {
    "dataschema": "buf.build/evalops/proto/events.v1.Change"
  }
}`),
	EventPipelineSignalCreateLinkedInActive: []byte(`{
  "spec_version": "1.0",
  "id": "evt_pipeline_signal_linkedin_active_1",
  "type": "pipeline.changes.signal.create",
  "source": "pipeline",
  "subject": "pipeline.changes.signal.create",
  "time": "2026-04-13T08:45:00.23456789Z",
  "data_content_type": "application/protobuf",
  "tenant_id": "11111111-1111-1111-1111-111111111111",
  "data": {
    "@type": "type.googleapis.com/events.v1.Change",
    "seq": "205",
    "organization_id": "11111111-1111-1111-1111-111111111111",
    "aggregate_type": "signal",
    "aggregate_id": "55555555-5555-5555-5555-555555555555",
    "operation": "create",
    "actor_type": "service",
    "actor_id": "pipeline-api",
    "aggregate_version": "1",
    "recorded_at": "2026-04-13T08:45:00.23456789Z",
    "payload": {
      "id": "55555555-5555-5555-5555-555555555555",
      "organization_id": "11111111-1111-1111-1111-111111111111",
      "owner_type": "contact",
      "owner_id": "44444444-4444-4444-4444-444444444444",
      "signal_type": "linkedin_active",
      "source": "linkedin",
      "strength": 87,
      "data": {
        "company_name": "Acme Corporation",
        "recent_title": "VP Sales"
      },
      "created_at": "2026-04-13T08:45:00.23456789Z",
      "updated_at": "2026-04-13T08:45:00.23456789Z"
    }
  },
  "extensions": {
    "dataschema": "buf.build/evalops/proto/events.v1.Change"
  }
}`),
	MeterRecordUsageRequestLLMGatewayResponses: []byte(`{
  "team_id": "team_eng",
  "agent_id": "agent_456",
  "surface": "chat",
  "event_type": "llm.completion",
  "model": "gpt-5.4",
  "provider": "openai",
  "input_tokens": "100",
  "output_tokens": "50",
  "cache_read_tokens": "0",
  "cache_write_tokens": "0",
  "total_cost_usd": 0.001,
  "request_id": "resp_meter",
  "metadata": {
    "agent_id": "agent_456",
    "surface": "chat",
    "pipeline_deal_id": "deal_123",
    "pipeline_sequence_id": "seq_456",
    "endpoint": "/v1/responses",
    "provider_ref_id": "pref_000001",
    "pricing_status": "gateway_estimated"
  },
  "data": {
    "endpoint": "/v1/responses",
    "provider_ref_id": "pref_000001"
  }
}`),
	KeysResolveProviderRefRequest: []byte(`{
  "provider": "openai",
  "environment": "production",
  "credential_name": "default",
  "team_id": "team_platform"
}`),
	KeysResolveProviderRefResponse: []byte(`{
  "provider_ref": {
    "id": "pref_000001",
    "organization_id": "org_123",
    "provider": "openai",
    "environment": "production",
    "credential_name": "default",
    "team_id": "team_platform",
    "credential_type": "api_key",
    "endpoint_url": "https://api.openai.com/v1",
    "region": "us",
    "credential_data": {
      "api_key": "sk-live-123"
    },
    "provider_config": {
      "base_url": "https://api.openai.com/v1"
    },
    "created_at": "2026-04-13T11:30:00Z",
    "updated_at": "2026-04-13T11:30:00Z"
  }
}`),
}

// Catalog returns the canonical contract fixture paths exported by this module.
func Catalog() []string {
	return append([]string(nil), fixtureCatalog...)
}

// Read returns a canonical proto fixture from the proto fixture catalog.
// Callers can pass either a path rooted at proto/ or a service-relative path.
func Read(name string) ([]byte, error) {
	cleaned := path.Clean(strings.TrimSpace(name))
	if cleaned == "." || cleaned == "" {
		return nil, fmt.Errorf("fixture path is required")
	}
	if strings.HasPrefix(cleaned, "../") {
		return nil, fmt.Errorf("fixture path %q escapes the fixture catalog", name)
	}
	if data, ok := embeddedFixture(cleaned); ok {
		return data, nil
	}
	if !strings.HasPrefix(cleaned, "proto/") {
		cleaned = path.Join("proto", cleaned)
	}
	if data, ok := embeddedFixture(cleaned); ok {
		return data, nil
	}
	root, err := moduleRoot()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(cleaned)))
}

func embeddedFixture(name string) ([]byte, bool) {
	if data, ok := embeddedFixtures[name]; ok {
		return append([]byte(nil), data...), true
	}
	if strings.HasPrefix(name, "proto/") {
		trimmed := strings.TrimPrefix(name, "proto/")
		if data, ok := embeddedFixtures[trimmed]; ok {
			return append([]byte(nil), data...), true
		}
	}
	return nil, false
}

func moduleRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve contract fixture source path")
	}
	return filepath.Dir(filepath.Dir(file)), nil
}

func LoadCloudEvent(name string) (*eventsv1.CloudEvent, error) {
	data, err := Read(name)
	if err != nil {
		return nil, err
	}
	message, err := eventhelpers.UnmarshalCloudEventProtoJSON(data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal fixture %q: %w", name, err)
	}
	return message, nil
}

func LoadChangeFixture(name string) (*eventsv1.CloudEvent, *eventsv1.Change, error) {
	envelope, err := LoadCloudEvent(name)
	if err != nil {
		return nil, nil, err
	}
	change, err := eventhelpers.UnpackChange(envelope)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal change fixture %q: %w", name, err)
	}
	return envelope, change, nil
}

func LoadEvaluationCompletedFixture(name string) (*eventsv1.CloudEvent, *eventsv1.EvaluationCompleted, error) {
	envelope, err := LoadCloudEvent(name)
	if err != nil {
		return nil, nil, err
	}
	message, err := eventhelpers.UnpackEvaluationCompleted(envelope)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal evaluation fixture %q: %w", name, err)
	}
	return envelope, message, nil
}

func LoadTapFixture(name string) (*eventsv1.CloudEvent, *tapv1.TapEventData, error) {
	envelope, err := LoadCloudEvent(name)
	if err != nil {
		return nil, nil, err
	}
	data, err := eventhelpers.UnpackTapEventData(envelope)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal tap fixture %q: %w", name, err)
	}
	return envelope, data, nil
}

func LoadFeatureFlagSnapshot(name string) (*configv1.FeatureFlagSnapshot, error) {
	var message configv1.FeatureFlagSnapshot
	if err := UnmarshalProtoJSON(name, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

// UnmarshalProtoJSON reads a canonical fixture and unmarshals it with strict
// protojson parsing, rejecting unknown fields.
func UnmarshalProtoJSON(name string, message proto.Message) error {
	data, err := Read(name)
	if err != nil {
		return err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(data, message); err != nil {
		return fmt.Errorf("unmarshal fixture %q: %w", name, err)
	}
	return nil
}
