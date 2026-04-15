package eventhelpers

import (
	"errors"
	"testing"
	"time"

	eventsv1 "github.com/evalops/proto/gen/go/events/v1"
	tapv1 "github.com/evalops/proto/gen/go/tap/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNewCloudEventRoundTripPreservesTypedChangePayload(t *testing.T) {
	t.Parallel()

	change := &eventsv1.Change{
		Seq:            42,
		OrganizationId: "11111111-1111-1111-1111-111111111111",
		AggregateType:  "activity",
		AggregateId:    "33333333-3333-3333-3333-333333333333",
		Operation:      "create",
		Payload: mustStruct(t, map[string]any{
			"outcome": "replied",
			"subject": "Re: Intro to EvalOps",
		}),
	}

	envelope, err := NewCloudEvent(
		"evt_123",
		"pipeline.changes.activity.create",
		"pipeline",
		"pipeline.changes.activity.create",
		change.GetOrganizationId(),
		time.Date(2026, 4, 13, 0, 30, 0, 123456789, time.UTC),
		change,
	)
	if err != nil {
		t.Fatalf("NewCloudEvent() error = %v", err)
	}

	payload, err := MarshalProtoJSON(envelope)
	if err != nil {
		t.Fatalf("MarshalProtoJSON() error = %v", err)
	}

	decoded, err := UnmarshalCloudEventProtoJSON(payload)
	if err != nil {
		t.Fatalf("UnmarshalCloudEventProtoJSON() error = %v", err)
	}
	if decoded.GetSpecVersion() != "1.0" {
		t.Fatalf("unexpected spec_version %q", decoded.GetSpecVersion())
	}
	if decoded.GetDataContentType() != CanonicalDataContentType {
		t.Fatalf("unexpected data_content_type %q", decoded.GetDataContentType())
	}
	if decoded.GetSubject() != "pipeline.changes.activity.create" {
		t.Fatalf("unexpected subject %q", decoded.GetSubject())
	}
	if got := decoded.GetExtensions()["dataschema"].GetStringValue(); got != "buf.build/evalops/proto/events.v1.Change" {
		t.Fatalf("unexpected dataschema %q", got)
	}

	unpacked, err := UnpackChange(decoded)
	if err != nil {
		t.Fatalf("UnpackChange() error = %v", err)
	}
	if unpacked.GetPayload().GetFields()["outcome"].GetStringValue() != "replied" {
		t.Fatalf("unexpected payload %#v", unpacked.GetPayload().AsMap())
	}
}

func TestUnpackTapEventDataRoundTrip(t *testing.T) {
	t.Parallel()

	data := &tapv1.TapEventData{
		Provider:   "hubspot",
		TenantId:   "11111111-1111-1111-1111-111111111111",
		EntityType: "deal",
		EntityId:   "deal_123",
		Changes: map[string]*tapv1.FieldChange{
			"stage": {
				From: structpb.NewStringValue("new"),
				To:   structpb.NewStringValue("qualified"),
			},
		},
	}

	envelope, err := NewCloudEvent(
		"evt_456",
		"siphon.tap.hubspot.deal.updated",
		"siphon",
		"hubspot/deal/deal_123",
		data.GetTenantId(),
		time.Date(2026, 4, 13, 1, 0, 0, 0, time.UTC),
		data,
	)
	if err != nil {
		t.Fatalf("NewCloudEvent() error = %v", err)
	}

	unpacked, err := UnpackTapEventData(envelope)
	if err != nil {
		t.Fatalf("UnpackTapEventData() error = %v", err)
	}
	if unpacked.GetProvider() != "hubspot" {
		t.Fatalf("unexpected provider %q", unpacked.GetProvider())
	}
	if unpacked.GetChanges()["stage"].GetTo().GetStringValue() != "qualified" {
		t.Fatalf("unexpected stage change %#v", unpacked.GetChanges()["stage"].GetTo())
	}
	if got := envelope.GetExtensions()["dataschema"].GetStringValue(); got != "buf.build/evalops/proto/tap.v1.TapEventData" {
		t.Fatalf("unexpected dataschema %q", got)
	}
}

func TestUnpackEvaluationCompletedRoundTrip(t *testing.T) {
	t.Parallel()

	message := &eventsv1.EvaluationCompleted{
		SignalType:     "technical_capability",
		Summary:        "Claude beats GPT-4 on latency by 40% on enterprise workloads.",
		SuccessRate:    float64ptr(0.9),
		CompanyDomains: []string{"acme.com"},
		CompanyNames:   []string{"Acme Corporation"},
		DealIds:        []string{"deal-123"},
		Run: &eventsv1.EvaluationRun{
			Id:            "run-1",
			TestSuiteId:   "suite-1",
			TestSuiteName: "Latency benchmark",
			Name:          "Enterprise latency proof",
			Description:   "Claude beats GPT-4 on latency by 40% on enterprise workloads.",
			Tags:          []string{"pipeline:signal_type=technical_capability"},
		},
		Metrics: &eventsv1.EvaluationMetrics{
			TotalTests:  int64ptr(20),
			PassedTests: int64ptr(18),
			FailedTests: int64ptr(2),
			TotalCost:   float64ptr(4.2),
			Duration:    float64ptr(12.5),
			SuccessRate: float64ptr(0.9),
		},
	}

	envelope, err := NewCloudEvent(
		"evt_eval_123",
		"evaluation.completed",
		"fermata",
		"product.evaluation.completed",
		"11111111-1111-1111-1111-111111111111",
		time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC),
		message,
	)
	if err != nil {
		t.Fatalf("NewCloudEvent() error = %v", err)
	}

	unpacked, err := UnpackEvaluationCompleted(envelope)
	if err != nil {
		t.Fatalf("UnpackEvaluationCompleted() error = %v", err)
	}
	if unpacked.GetSignalType() != "technical_capability" {
		t.Fatalf("unexpected signal_type %q", unpacked.GetSignalType())
	}
	if unpacked.GetMetrics().GetSuccessRate() != 0.9 {
		t.Fatalf("unexpected success_rate %v", unpacked.GetMetrics().GetSuccessRate())
	}
	if got := envelope.GetExtensions()["dataschema"].GetStringValue(); got != "buf.build/evalops/proto/events.v1.EvaluationCompleted" {
		t.Fatalf("unexpected dataschema %q", got)
	}
}

func TestNewChangeBuildsCanonicalMessageFromJSONPayload(t *testing.T) {
	t.Parallel()

	change, err := NewChange(
		108,
		"11111111-1111-1111-1111-111111111111",
		"activity",
		"33333333-3333-3333-3333-333333333333",
		"create",
		"service",
		"pipeline-api",
		1,
		time.Date(2026, 4, 13, 0, 30, 0, 123456789, time.UTC),
		[]byte(`{"outcome":"replied","channel":"email"}`),
	)
	if err != nil {
		t.Fatalf("NewChange() error = %v", err)
	}

	if change.GetSeq() != 108 {
		t.Fatalf("unexpected seq %d", change.GetSeq())
	}
	if change.GetRecordedAt().AsTime() != time.Date(2026, 4, 13, 0, 30, 0, 123456789, time.UTC) {
		t.Fatalf("unexpected recorded_at %s", change.GetRecordedAt().AsTime())
	}
	if change.GetPayload().GetFields()["outcome"].GetStringValue() != "replied" {
		t.Fatalf("unexpected payload %#v", change.GetPayload().AsMap())
	}
}

func TestNewChangeRejectsNonObjectPayload(t *testing.T) {
	t.Parallel()

	if _, err := NewChange(1, "org", "activity", "agg", "create", "service", "svc", 1, time.Time{}, []byte(`["bad"]`)); err == nil {
		t.Fatal("expected non-object payload error")
	}
}

func TestUnpackDataReturnsErrorWhenEnvelopeDataMissing(t *testing.T) {
	t.Parallel()

	err := UnpackData(&eventsv1.CloudEvent{}, &eventsv1.Change{})
	if !errors.Is(err, errDataNil) {
		t.Fatalf("UnpackData() error = %v, want %v", err, errDataNil)
	}
}

func TestUnpackChangeReturnsErrorWhenEnvelopeDataMissing(t *testing.T) {
	t.Parallel()

	_, err := UnpackChange(&eventsv1.CloudEvent{})
	if !errors.Is(err, errDataNil) {
		t.Fatalf("UnpackChange() error = %v, want %v", err, errDataNil)
	}
}

func TestUnpackTapEventDataReturnsErrorWhenEnvelopeDataMissing(t *testing.T) {
	t.Parallel()

	_, err := UnpackTapEventData(&eventsv1.CloudEvent{})
	if !errors.Is(err, errDataNil) {
		t.Fatalf("UnpackTapEventData() error = %v, want %v", err, errDataNil)
	}
}

func TestUnpackEvaluationCompletedReturnsErrorWhenEnvelopeDataMissing(t *testing.T) {
	t.Parallel()

	_, err := UnpackEvaluationCompleted(&eventsv1.CloudEvent{})
	if !errors.Is(err, errDataNil) {
		t.Fatalf("UnpackEvaluationCompleted() error = %v, want %v", err, errDataNil)
	}
}

func mustStruct(t *testing.T, fields map[string]any) *structpb.Struct {
	t.Helper()

	message, err := structpb.NewStruct(fields)
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}
	return message
}

func int64ptr(value int64) *int64 {
	return &value
}

func float64ptr(value float64) *float64 {
	return &value
}
