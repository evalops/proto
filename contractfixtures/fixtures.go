package contractfixtures

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	eventsv1 "github.com/evalops/proto/gen/go/events/v1"
	tapv1 "github.com/evalops/proto/gen/go/tap/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	EventPipelineActivityCreateReplied          = "events/v1/testdata/cloud_event_pipeline_activity_create_replied.json"
	EventParkerWorkRelationshipUpdateTerminated = "events/v1/testdata/cloud_event_parker_work_relationship_update_terminated.json"
	EventTapHubspotDealQualified                = "events/v1/testdata/cloud_event_tap_hubspot_deal_qualified.json"
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
	var message eventsv1.CloudEvent
	if err := unmarshalProtoJSONFixture(name, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

func LoadChangeFixture(name string) (*eventsv1.CloudEvent, *eventsv1.Change, error) {
	envelope, err := LoadCloudEvent(name)
	if err != nil {
		return nil, nil, err
	}
	var change eventsv1.Change
	if err := envelope.GetData().UnmarshalTo(&change); err != nil {
		return nil, nil, fmt.Errorf("unmarshal change fixture %q: %w", name, err)
	}
	return envelope, &change, nil
}

func LoadTapFixture(name string) (*eventsv1.CloudEvent, *tapv1.TapEventData, error) {
	envelope, err := LoadCloudEvent(name)
	if err != nil {
		return nil, nil, err
	}
	var data tapv1.TapEventData
	if err := envelope.GetData().UnmarshalTo(&data); err != nil {
		return nil, nil, fmt.Errorf("unmarshal tap fixture %q: %w", name, err)
	}
	return envelope, &data, nil
}

func unmarshalProtoJSONFixture(name string, message proto.Message) error {
	data, err := Read(name)
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(data, message); err != nil {
		return fmt.Errorf("unmarshal fixture %q: %w", name, err)
	}
	return nil
}
