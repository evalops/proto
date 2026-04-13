package contractfixtures

import (
	"bytes"
	"testing"

	meterv1 "github.com/evalops/proto/gen/go/meter/v1"
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
	if envelope.GetType() != "ensemble.tap.hubspot.deal.updated" {
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
