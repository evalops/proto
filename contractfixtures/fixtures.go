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
	ConfigFeatureFlagSnapshot                   = "config/v1/testdata/feature_flag_snapshot.json"
	EventPipelineActivityCreateReplied          = "events/v1/testdata/cloud_event_pipeline_activity_create_replied.json"
	EventParkerWorkRelationshipUpdateTerminated = "events/v1/testdata/cloud_event_parker_work_relationship_update_terminated.json"
	EventTapHubspotDealQualified                = "events/v1/testdata/cloud_event_tap_hubspot_deal_qualified.json"
	MeterRecordUsageRequestLLMGatewayResponses  = "meter/v1/testdata/record_usage_request_llm_gateway_responses.json"
	MeterRecordUsageRequest                     = "meter/v1/testdata/record_usage_request.json"
	MeterRecordUsageResponse                    = "meter/v1/testdata/record_usage_response.json"
	MeterUsageQueryResponse                     = "meter/v1/testdata/query_usage_response.json"
	MeterUsageSummaryResponse                   = "meter/v1/testdata/usage_summary_response.json"
	MeterMeterSummaryResponse                   = "meter/v1/testdata/meter_summary_response.json"
)

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
	if !strings.HasPrefix(cleaned, "proto/") {
		cleaned = path.Join("proto", cleaned)
	}
	root, err := moduleRoot()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(cleaned)))
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
