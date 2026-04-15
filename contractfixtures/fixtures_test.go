package contractfixtures

import (
	"bytes"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	configv1 "github.com/evalops/proto/gen/go/config/v1"
	eventsv1 "github.com/evalops/proto/gen/go/events/v1"
	keysv1 "github.com/evalops/proto/gen/go/keys/v1"
	meterv1 "github.com/evalops/proto/gen/go/meter/v1"
	workflowsv1 "github.com/evalops/proto/gen/go/workflows/v1"
)

func TestReadSupportsRelativeFixturePaths(t *testing.T) {
	t.Parallel()

	data, err := Read(EventPipelineActivityCreateReplied)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if !bytes.Contains(data, []byte(`"type": "pipeline.changes.activity.create"`)) {
		t.Fatalf("expected pipeline activity fixture, got %s", string(data))
	}
}

func TestReadSupportsProtoPrefixedFixturePaths(t *testing.T) {
	t.Parallel()

	data, err := Read("proto/" + EventParkerWorkRelationshipUpdateTerminated)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if !bytes.Contains(data, []byte(`"type": "parker.changes.work_relationship.update"`)) {
		t.Fatalf("expected parker work relationship fixture, got %s", string(data))
	}
}

func TestLoadChangeFixtureSupportsPipelineSignalFixture(t *testing.T) {
	t.Parallel()

	envelope, message, err := LoadChangeFixture(EventPipelineSignalCreateLinkedInActive)
	if err != nil {
		t.Fatalf("load pipeline signal fixture: %v", err)
	}
	if envelope.GetType() != "pipeline.changes.signal.create" {
		t.Fatalf("unexpected type %q", envelope.GetType())
	}
	if message.GetAggregateType() != "signal" || message.GetOperation() != "create" {
		t.Fatalf("unexpected aggregate/operation %q/%q", message.GetAggregateType(), message.GetOperation())
	}
	if got := message.GetPayload().AsMap()["signal_type"]; got != "linkedin_active" {
		t.Fatalf("unexpected signal_type %#v", got)
	}
}

func TestLoadChangeFixtureSupportsPipelineDealFixture(t *testing.T) {
	t.Parallel()

	envelope, message, err := LoadChangeFixture(EventPipelineDealUpdateClosedWon)
	if err != nil {
		t.Fatalf("load pipeline deal fixture: %v", err)
	}
	if envelope.GetType() != "pipeline.changes.deal.update" {
		t.Fatalf("unexpected type %q", envelope.GetType())
	}
	if message.GetAggregateType() != "deal" || message.GetOperation() != "update" {
		t.Fatalf("unexpected aggregate/operation %q/%q", message.GetAggregateType(), message.GetOperation())
	}
	if got := message.GetPayload().AsMap()["stage"]; got != "closed_won" {
		t.Fatalf("unexpected stage %#v", got)
	}
}

func TestLoadWorkflowRunLifecycleEvent(t *testing.T) {
	t.Parallel()

	message, err := LoadWorkflowRunLifecycleEvent(WorkflowsRunLifecycleEventStarted)
	if err != nil {
		t.Fatalf("load workflow run lifecycle event: %v", err)
	}
	if message.GetRun().GetId() != "wfr_outbound_acme_001" {
		t.Fatalf("unexpected run.id %q", message.GetRun().GetId())
	}
	if message.GetRun().GetState() != workflowsv1.WorkflowState_WORKFLOW_STATE_RUNNING {
		t.Fatalf("unexpected run.state %s", message.GetRun().GetState())
	}
	if len(message.GetRun().GetSteps()) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(message.GetRun().GetSteps()))
	}
}

func TestLoadWorkflowStepLifecycleEvent(t *testing.T) {
	t.Parallel()

	message, err := LoadWorkflowStepLifecycleEvent(WorkflowsStepLifecycleEventWaiting)
	if err != nil {
		t.Fatalf("load workflow step lifecycle event: %v", err)
	}
	if message.GetRunId() != "wfr_outbound_acme_001" {
		t.Fatalf("unexpected run_id %q", message.GetRunId())
	}
	if message.GetWorkflowState() != workflowsv1.WorkflowState_WORKFLOW_STATE_RUNNING {
		t.Fatalf("unexpected workflow_state %s", message.GetWorkflowState())
	}
	if message.GetStep().GetState() != workflowsv1.StepRunState_STEP_RUN_STATE_WAITING {
		t.Fatalf("unexpected step.state %s", message.GetStep().GetState())
	}
}

func TestReadRejectsEscapingPaths(t *testing.T) {
	t.Parallel()

	if _, err := Read("../secrets.txt"); err == nil {
		t.Fatal("expected escaping path to fail")
	}
}

func TestLoadTapFixture(t *testing.T) {
	t.Parallel()

	envelope, data, err := LoadTapFixture(EventTapHubspotDealQualified)
	if err != nil {
		t.Fatalf("load tap fixture: %v", err)
	}
	if envelope.GetType() != "siphon.hubspot.deal.updated" {
		t.Fatalf("unexpected type %q", envelope.GetType())
	}
	if envelope.GetTenantId() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected tenant_id %q", envelope.GetTenantId())
	}
	if data.GetProvider() != "hubspot" {
		t.Fatalf("unexpected provider %q", data.GetProvider())
	}
	if data.GetChanges()["stage"].GetTo().GetStringValue() != "qualified" {
		t.Fatalf("unexpected stage change %#v", data.GetChanges()["stage"].GetTo())
	}
}

func TestLoadEvaluationCompletedFixture(t *testing.T) {
	t.Parallel()

	envelope, message, err := LoadEvaluationCompletedFixture(EventEvaluationCompletedTechnicalCapability)
	if err != nil {
		t.Fatalf("load evaluation fixture: %v", err)
	}
	if envelope.GetType() != "evaluation.completed" {
		t.Fatalf("unexpected type %q", envelope.GetType())
	}
	if got := envelope.GetExtensions()["dataschema"].GetStringValue(); got != "buf.build/evalops/proto/events.v1.EvaluationCompleted" {
		t.Fatalf("unexpected dataschema %q", got)
	}
	if message.GetSignalType() != "technical_capability" {
		t.Fatalf("unexpected signal_type %q", message.GetSignalType())
	}
	if message.GetRun().GetId() != "run-1" {
		t.Fatalf("unexpected run.id %q", message.GetRun().GetId())
	}
	if message.GetMetrics().GetSuccessRate() != 0.9 {
		t.Fatalf("unexpected metrics.success_rate %v", message.GetMetrics().GetSuccessRate())
	}
	if envelope.GetData().GetTypeUrl() != "type.googleapis.com/events.v1.EvaluationCompleted" {
		t.Fatalf("unexpected type URL %q", envelope.GetData().GetTypeUrl())
	}
	var direct eventsv1.EvaluationCompleted
	if err := envelope.GetData().UnmarshalTo(&direct); err != nil {
		t.Fatalf("unmarshal direct evaluation payload: %v", err)
	}
}

func TestUnmarshalProtoJSONLoadsMeterFixture(t *testing.T) {
	t.Parallel()

	var message meterv1.RecordUsageRequest
	if err := UnmarshalProtoJSON(MeterRecordUsageRequest, &message); err != nil {
		t.Fatalf("load meter record usage request: %v", err)
	}
	if message.GetEventType() != "llm.completion" {
		t.Fatalf("unexpected event_type %q", message.GetEventType())
	}
	if message.GetMetadata().GetFields()["pipeline_deal_id"].GetStringValue() != "deal_123" {
		t.Fatalf("unexpected pipeline_deal_id %#v", message.GetMetadata().GetFields()["pipeline_deal_id"])
	}
}

func TestUnmarshalProtoJSONLoadsLLMGatewayMeterFixture(t *testing.T) {
	t.Parallel()

	var message meterv1.RecordUsageRequest
	if err := UnmarshalProtoJSON(MeterRecordUsageRequestLLMGatewayResponses, &message); err != nil {
		t.Fatalf("load llm-gateway meter record usage request: %v", err)
	}
	if message.GetEventType() != "llm.completion" {
		t.Fatalf("unexpected event_type %q", message.GetEventType())
	}
	if message.GetRequestId() != "resp_meter" {
		t.Fatalf("unexpected request_id %q", message.GetRequestId())
	}
	if message.GetData().GetFields()["endpoint"].GetStringValue() != "/v1/responses" {
		t.Fatalf("unexpected endpoint %#v", message.GetData().GetFields()["endpoint"])
	}
}

func TestUnmarshalProtoJSONLoadsKeysResolveRequestFixture(t *testing.T) {
	t.Parallel()

	var message keysv1.ResolveProviderRefRequest
	if err := UnmarshalProtoJSON(KeysResolveProviderRefRequest, &message); err != nil {
		t.Fatalf("load keys resolve provider ref request: %v", err)
	}
	if message.GetProvider() != "openai" {
		t.Fatalf("unexpected provider %q", message.GetProvider())
	}
	if message.GetTeamId() != "team_platform" {
		t.Fatalf("unexpected team_id %q", message.GetTeamId())
	}
}

func TestUnmarshalProtoJSONLoadsKeysResolveFixture(t *testing.T) {
	t.Parallel()

	var message keysv1.ResolveProviderRefResponse
	if err := UnmarshalProtoJSON(KeysResolveProviderRefResponse, &message); err != nil {
		t.Fatalf("load keys resolve provider ref response: %v", err)
	}
	if message.GetProviderRef().GetId() != "pref_000001" {
		t.Fatalf("unexpected provider_ref.id %q", message.GetProviderRef().GetId())
	}
	if message.GetProviderRef().GetCredentialData().GetFields()["api_key"].GetStringValue() != "sk-live-123" {
		t.Fatalf("unexpected credential_data api_key %#v", message.GetProviderRef().GetCredentialData().GetFields()["api_key"])
	}
}

func TestLoadFeatureFlagSnapshot(t *testing.T) {
	t.Parallel()

	message, err := LoadFeatureFlagSnapshot(ConfigFeatureFlagSnapshot)
	if err != nil {
		t.Fatalf("load feature flag snapshot: %v", err)
	}
	if message.GetSchemaVersion() != 1 {
		t.Fatalf("unexpected schema_version %d", message.GetSchemaVersion())
	}
	if len(message.GetFlags()) != 12 {
		t.Fatalf("expected 12 flags, got %d", len(message.GetFlags()))
	}
	if message.GetFlags()[0].GetKey() != "llm_gateway.model_routing.provider_failover" {
		t.Fatalf("unexpected first flag key %q", message.GetFlags()[0].GetKey())
	}
	if message.GetFlags()[6].GetKey() != "platform.kill_switches.gate.control_api" {
		t.Fatalf("unexpected seventh flag key %q", message.GetFlags()[6].GetKey())
	}
	if message.GetFlags()[8].GetKey() != "platform.kill_switches.prompts.resolve_api" {
		t.Fatalf("unexpected ninth flag key %q", message.GetFlags()[8].GetKey())
	}
	if message.GetFlags()[11].GetKey() != "platform.kill_switches.dagster.dbt_analytics_schedule" {
		t.Fatalf("unexpected twelfth flag key %q", message.GetFlags()[11].GetKey())
	}
}

func TestUnmarshalProtoJSONLoadsConfigFixture(t *testing.T) {
	t.Parallel()

	var message configv1.FeatureFlagSnapshot
	if err := UnmarshalProtoJSON(ConfigFeatureFlagSnapshot, &message); err != nil {
		t.Fatalf("load config feature flag snapshot: %v", err)
	}
	if message.GetFlags()[2].GetRolloutPercent() != 0 {
		t.Fatalf("unexpected third rollout_percent %d", message.GetFlags()[2].GetRolloutPercent())
	}
}

func TestCatalogMatchesFixtureTree(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob(filepath.Join("..", "proto", "*", "v1", "testdata", "*.json"))
	if err != nil {
		t.Fatalf("glob fixture tree: %v", err)
	}

	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		rel, err := filepath.Rel(filepath.Join("..", "proto"), match)
		if err != nil {
			t.Fatalf("rel fixture path: %v", err)
		}
		paths = append(paths, filepath.ToSlash(rel))
	}

	catalog := Catalog()
	sort.Strings(paths)
	sort.Strings(catalog)

	if !reflect.DeepEqual(catalog, paths) {
		t.Fatalf("fixture catalog drifted:\nwant %v\n got %v", paths, catalog)
	}
}

func TestCatalogFixturesAreReadable(t *testing.T) {
	t.Parallel()

	for _, fixture := range Catalog() {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			data, err := Read(fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("expected fixture bytes")
			}
		})
	}
}
