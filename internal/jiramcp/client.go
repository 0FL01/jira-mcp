package jiramcp

import (
	"context"
	"encoding/json"

	"github.com/mmatczuk/jira-mcp/internal/jira"
)

// JiraClient defines the Jira operations used by the MCP handlers.
// All operations are compatible with Jira Server/Data Center 7.x (REST API v2).
// NOTE: Rich text (ADF) and field_options are NOT supported.
type JiraClient interface {
	GetIssue(ctx context.Context, key string, opts *jira.GetQueryOptions) (*jira.Issue, error)
	// SearchIssues uses v2 API with offset-based pagination (startAt/maxResults)
	SearchIssues(ctx context.Context, jql string, opts *jira.SearchOptionsV2) (*jira.SearchResultV2, error)
	// CreateIssueV2 uses v2 API (Jira Server 7.x compatible)
	CreateIssueV2(ctx context.Context, payload map[string]any) (string, string, error)
	// UpdateIssueV2 uses v2 API (Jira Server 7.x compatible)
	UpdateIssueV2(ctx context.Context, key string, payload map[string]any) error
	DeleteIssue(ctx context.Context, key string) error
	DoTransition(ctx context.Context, key, transitionID string) error
	// AddComment uses v2 API with plain text body (ADF NOT supported)
	AddComment(ctx context.Context, key string, body string) (string, error)
	// UpdateComment uses v2 API with plain text body (ADF NOT supported)
	UpdateComment(ctx context.Context, key, commentID string, body string) error
	GetAllBoards(ctx context.Context, opts *jira.BoardListOptions) ([]jira.Board, bool, error)
	GetAllSprints(ctx context.Context, boardID int, opts *jira.GetAllSprintsOptions) ([]jira.Sprint, bool, error)
	GetSprintIssues(ctx context.Context, sprintID int) ([]jira.Issue, error)
	MoveIssuesToSprint(ctx context.Context, sprintID int, issueKeys []string) error
	GetAllProjects(ctx context.Context) (*jira.ProjectList, error)
	GetFields(ctx context.Context) ([]jira.Field, error)
	GetTransitions(ctx context.Context, key string) ([]jira.Transition, error)
	// GetFieldOptions is NOT supported on Jira Server 7.x - returns error
	GetFieldOptions(ctx context.Context, fieldID string) ([]json.RawMessage, error)
}
