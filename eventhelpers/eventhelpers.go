package eventhelpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	eventsv1 "github.com/evalops/proto/gen/go/events/v1"
	tapv1 "github.com/evalops/proto/gen/go/tap/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const CanonicalDataContentType = "application/protobuf"

var (
	errMessageNil  = errors.New("message_nil")
	errEnvelopeNil = errors.New("envelope_nil")
	errTargetNil   = errors.New("target_nil")
)

var marshalOptions = protojson.MarshalOptions{UseProtoNames: true}
var unmarshalOptions = protojson.UnmarshalOptions{DiscardUnknown: false}

// MarshalProtoJSON marshals a protobuf message with stable proto field names.
func MarshalProtoJSON(message proto.Message) ([]byte, error) {
	if message == nil {
		return nil, errMessageNil
	}
	return marshalOptions.Marshal(message)
}

// UnmarshalProtoJSON unmarshals protobuf JSON without discarding unknown fields.
func UnmarshalProtoJSON(data []byte, message proto.Message) error {
	if message == nil {
		return errTargetNil
	}
	if err := unmarshalOptions.Unmarshal(data, message); err != nil {
		return err
	}
	return nil
}

// UnmarshalCloudEventProtoJSON unmarshals a canonical events/v1.CloudEvent envelope.
func UnmarshalCloudEventProtoJSON(data []byte) (*eventsv1.CloudEvent, error) {
	var envelope eventsv1.CloudEvent
	if err := UnmarshalProtoJSON(data, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

// NewCloudEvent builds the canonical events/v1.CloudEvent envelope around a typed payload.
func NewCloudEvent(id, eventType, source, subject, tenantID string, occurredAt time.Time, payload proto.Message) (*eventsv1.CloudEvent, error) {
	if payload == nil {
		return nil, errMessageNil
	}
	anyPayload, err := anypb.New(payload)
	if err != nil {
		return nil, fmt.Errorf("pack payload: %w", err)
	}

	envelope := &eventsv1.CloudEvent{
		SpecVersion:     "1.0",
		Id:              strings.TrimSpace(id),
		Type:            strings.TrimSpace(eventType),
		Source:          strings.TrimSpace(source),
		Subject:         strings.TrimSpace(subject),
		DataContentType: CanonicalDataContentType,
		TenantId:        strings.TrimSpace(tenantID),
		Data:            anyPayload,
		Extensions: map[string]*structpb.Value{
			"dataschema": structpb.NewStringValue(dataschemaFor(payload)),
		},
	}
	if !occurredAt.IsZero() {
		envelope.Time = timestamppb.New(occurredAt.UTC())
	}
	return envelope, nil
}

// NewChange converts a canonical change payload into the shared events/v1.Change message.
func NewChange(
	seq int64,
	organizationID string,
	aggregateType string,
	aggregateID string,
	operation string,
	actorType string,
	actorID string,
	aggregateVersion int64,
	recordedAt time.Time,
	rawPayload []byte,
) (*eventsv1.Change, error) {
	payload, err := payloadStructFromJSON(rawPayload)
	if err != nil {
		return nil, err
	}

	var timestamp *timestamppb.Timestamp
	if !recordedAt.IsZero() {
		timestamp = timestamppb.New(recordedAt.UTC())
	}

	return &eventsv1.Change{
		Seq:              seq,
		OrganizationId:   strings.TrimSpace(organizationID),
		AggregateType:    strings.TrimSpace(aggregateType),
		AggregateId:      strings.TrimSpace(aggregateID),
		Operation:        strings.TrimSpace(operation),
		ActorType:        strings.TrimSpace(actorType),
		ActorId:          strings.TrimSpace(actorID),
		AggregateVersion: aggregateVersion,
		RecordedAt:       timestamp,
		Payload:          payload,
	}, nil
}

// UnpackData unmarshals the typed Any payload from a canonical CloudEvent.
func UnpackData(envelope *eventsv1.CloudEvent, target proto.Message) error {
	if envelope == nil {
		return errEnvelopeNil
	}
	if target == nil {
		return errTargetNil
	}
	if envelope.GetData() == nil {
		return nil
	}
	return envelope.GetData().UnmarshalTo(target)
}

func UnpackChange(envelope *eventsv1.CloudEvent) (*eventsv1.Change, error) {
	message := &eventsv1.Change{}
	if err := UnpackData(envelope, message); err != nil {
		return nil, err
	}
	return message, nil
}

func UnpackTapEventData(envelope *eventsv1.CloudEvent) (*tapv1.TapEventData, error) {
	message := &tapv1.TapEventData{}
	if err := UnpackData(envelope, message); err != nil {
		return nil, err
	}
	return message, nil
}

func dataschemaFor(payload proto.Message) string {
	return "buf.build/evalops/proto/" + string(payload.ProtoReflect().Descriptor().FullName())
}

func payloadStructFromJSON(rawPayload []byte) (*structpb.Struct, error) {
	trimmed := bytes.TrimSpace(rawPayload)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}

	var value any
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return nil, fmt.Errorf("unmarshal change payload: %w", err)
	}
	fields, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("change payload must be a JSON object")
	}

	message, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, fmt.Errorf("encode canonical change payload: %w", err)
	}
	return message, nil
}
