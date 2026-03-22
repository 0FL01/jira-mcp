package jiramcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/mmatczuk/jira-mcp/internal/jira"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func callSchema(t *testing.T, h *handlers, args SchemaArgs) (string, bool) {
	t.Helper()
	result, _, err := h.handleSchema(context.Background(), nil, args)
	require.NoError(t, err)
	text := result.Content[0].(*mcp.TextContent).Text
	return text, result.IsError
}

// --- dispatch ---

func TestHandleSchema_UnknownResource(t *testing.T) {
	h := &handlers{client: &mockClient{}}
	text, isErr := callSchema(t, h, SchemaArgs{Resource: "bogus"})
	assert.True(t, isErr)
	assert.Contains(t, text, `Unknown resource "bogus"`)
}

// --- fields ---

func TestSchemaFields_Success(t *testing.T) {
	mc := &mockClient{
		GetFieldsFn: func(context.Context) ([]jira.Field, error) {
			return []jira.Field{
				{
					ID:     "summary",
					Name:   "Summary",
					Custom: false,
					Schema: jira.FieldSchema{Type: "string"},
				},
				{
					ID:     "customfield_10001",
					Name:   "Story Points",
					Custom: true,
					Schema: jira.FieldSchema{Type: "number"},
				},
			}, nil
		},
	}
	h := &handlers{client: mc}
	text, isErr := callSchema(t, h, SchemaArgs{Resource: "fields"})
	assert.False(t, isErr)
	assert.Contains(t, text, "Found 2 field(s)")
	assert.Contains(t, text, "summary")
	assert.Contains(t, text, "customfield_10001")
	assert.Contains(t, text, "Story Points")
}

func TestSchemaFields_Error(t *testing.T) {
	mc := &mockClient{
		GetFieldsFn: func(context.Context) ([]jira.Field, error) {
			return nil, fmt.Errorf("auth expired")
		},
	}
	h := &handlers{client: mc}
	text, isErr := callSchema(t, h, SchemaArgs{Resource: "fields"})
	assert.True(t, isErr)
	assert.Contains(t, text, "auth expired")
}

func TestSchemaFields_SchemaItems(t *testing.T) {
	mc := &mockClient{
		GetFieldsFn: func(context.Context) ([]jira.Field, error) {
			return []jira.Field{
				{
					ID:   "labels",
					Name: "Labels",
					Schema: jira.FieldSchema{
						Type:  "array",
						Items: "string",
					},
				},
			}, nil
		},
	}
	h := &handlers{client: mc}
	text, _ := callSchema(t, h, SchemaArgs{Resource: "fields"})
	assert.Contains(t, text, "schema_items")
	assert.Contains(t, text, "string")
}

// --- transitions ---

func TestSchemaTransitions_Success(t *testing.T) {
	mc := &mockClient{
		GetTransitionsFn: func(_ context.Context, key string) ([]jira.Transition, error) {
			assert.Equal(t, "PROJ-1", key)
			return []jira.Transition{
				{ID: "11", Name: "Start Progress", To: jira.Status{Name: "In Progress"}},
				{ID: "21", Name: "Done", To: jira.Status{Name: "Done"}},
			}, nil
		},
	}
	h := &handlers{client: mc}
	text, isErr := callSchema(t, h, SchemaArgs{Resource: "transitions", IssueKey: "PROJ-1"})
	assert.False(t, isErr)
	assert.Contains(t, text, "Found 2 transition(s)")
	assert.Contains(t, text, "Start Progress")
	assert.Contains(t, text, "In Progress")
}

func TestSchemaTransitions_NoIssueKey(t *testing.T) {
	h := &handlers{client: &mockClient{}}
	text, isErr := callSchema(t, h, SchemaArgs{Resource: "transitions"})
	assert.True(t, isErr)
	assert.Contains(t, text, "issue_key is required")
}

func TestSchemaTransitions_Error(t *testing.T) {
	mc := &mockClient{
		GetTransitionsFn: func(context.Context, string) ([]jira.Transition, error) {
			return nil, fmt.Errorf("issue not found")
		},
	}
	h := &handlers{client: mc}
	text, isErr := callSchema(t, h, SchemaArgs{Resource: "transitions", IssueKey: "BAD-1"})
	assert.True(t, isErr)
	assert.Contains(t, text, "issue not found")
}

// --- field_options (NOT SUPPORTED on Jira Server 7.x) ---

// TestSchemaFieldOptions_Unsupported verifies that field_options returns
// a message indicating this feature is not supported on Jira Server 7.x.
func TestSchemaFieldOptions_Unsupported(t *testing.T) {
	h := &handlers{client: &mockClient{}}
	text, isErr := callSchema(t, h, SchemaArgs{Resource: "field_options", FieldID: "customfield_10001"})
	// field_options is now a warning, not an error
	assert.False(t, isErr)
	assert.Contains(t, text, "NOT supported on Jira Server 7.x")
	assert.Contains(t, text, "REST API v3")
}

// TestSchemaFieldOptions_NoFieldIDStillUnsupported verifies that even without
// field_id, the response indicates field_options is unsupported.
func TestSchemaFieldOptions_NoFieldIDStillUnsupported(t *testing.T) {
	h := &handlers{client: &mockClient{}}
	text, isErr := callSchema(t, h, SchemaArgs{Resource: "field_options"})
	// Even without field_id, we show the unsupported message
	assert.False(t, isErr)
	assert.Contains(t, text, "NOT supported on Jira Server 7.x")
}
