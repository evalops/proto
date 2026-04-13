package contractfixtures

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	EventPipelineActivityCreateReplied          = "events/v1/testdata/cloud_event_pipeline_activity_create_replied.json"
	EventParkerWorkRelationshipUpdateTerminated = "events/v1/testdata/cloud_event_parker_work_relationship_update_terminated.json"
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
