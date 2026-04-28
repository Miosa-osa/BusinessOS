package adapters

import (
	"context"
	"strings"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type salesforceAdapter struct{}

func init() { connectors.Register(&salesforceAdapter{}) }

func (s *salesforceAdapter) Kind() string                      { return "salesforce" }
func (s *salesforceAdapter) DisplayName() string               { return "Salesforce" }
func (s *salesforceAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (s *salesforceAdapter) RequiredConfigKeys() []string      { return []string{"instance_url"} }

type salesforceState struct {
	instanceURL string
	accessToken string
}

func (s *salesforceAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, s.RequiredConfigKeys()); err != nil {
		return nil, err
	}
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &salesforceState{
		instanceURL: connectors.StringField(cfg, "instance_url"),
		accessToken: connectors.CredentialField(cfg, "access_token"),
	}, nil
}

func (s *salesforceAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform reads any sObject (Account, Contact, Opportunity, Lead, …).
// Genre is the lower-cased sObject type, so downstream classifiers can
// route Salesforce records to the right node by sObject.
func (s *salesforceAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "Id")
	title := firstString(raw, "Name", "Subject")
	if title == "" {
		title = extID
	}
	body := firstString(raw, "Description", "Body")

	sobject := nestedString(raw, "attributes", "type")
	if sobject == "" {
		sobject = "record"
	}
	genre := strings.ToLower(sobject)

	sig := connectors.NewSignal(
		connectors.SignalID("salesforce", extID),
		title,
		body,
		connectors.SourceURI("salesforce", extID),
		genre,
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "LastModifiedDate"))
	return sig, nil
}
