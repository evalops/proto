package contractfixtures

import (
	"bytes"
	"testing"
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
