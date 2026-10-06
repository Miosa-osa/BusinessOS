package templates

import (
	"strings"
	"testing"
)

func renderWorkspaceInvitationForTest(t *testing.T, data WorkspaceInvitationData) (string, string) {
	t.Helper()

	service := &EmailTemplateService{
		appURL:       "https://app.businessos.dev",
		logoURL:      "https://storage.googleapis.com/businessos-downloads/email-icon.png",
		supportEmail: "roberto@miosa.ai",
	}
	html, plainText, err := service.RenderWorkspaceInvitation(data)
	if err != nil {
		t.Fatalf("render invitation template: %v", err)
	}
	return html, plainText
}

func TestWorkspaceInvitationTemplateHasMobileLayoutGuards(t *testing.T) {
	html, plainText := renderWorkspaceInvitationForTest(t, WorkspaceInvitationData{
		InviterName:    "Roberto H Luna",
		InviterEmail:   "roberto@miosa.ai",
		WorkspaceName:  "Northstar Growth",
		Role:           "member",
		InvitationLink: "https://app.businessos.dev/invite/403e3e0547dbcb2c6a07f8ef14e9c5237247a9c229649a81f771341bd353e4ea",
		ExpiresAt:      "on September 15, 2026",
	})

	required := []string{
		"* { box-sizing: border-box; }",
		"table-layout: fixed",
		"overflow-wrap: anywhere",
		"class=\"detail-cell\"",
		"display: block !important; width: 100% !important",
		"class=\"button\"",
	}
	for _, fragment := range required {
		if !strings.Contains(html, fragment) {
			t.Errorf("rendered invitation is missing mobile guard %q", fragment)
		}
	}

	if strings.Contains(html, "&mdash;") || strings.Contains(html, "—") {
		t.Fatal("rendered invitation must not contain an em dash")
	}
	if !strings.Contains(plainText, "Accept invitation") || !strings.Contains(plainText, "September 15, 2026") {
		t.Fatal("plain-text invitation is missing its action or exact expiration")
	}
}

func TestWorkspaceInvitationTemplateHandlesLongAndMissingValues(t *testing.T) {
	html, _ := renderWorkspaceInvitationForTest(t, WorkspaceInvitationData{
		InviterEmail:   "a-very-long-inviter-address@an-unusually-long-company-domain.example",
		WorkspaceName:  "Agency Operations and International Client Delivery Workspace",
		InvitationLink: "https://app.businessos.dev/invite/" + strings.Repeat("a", 160),
		ExpiresAt:      "on September 15, 2026",
	})

	if !strings.Contains(html, "a-very-long-inviter-address@an-unusually-long-company-domain.example invited you") {
		t.Fatal("missing inviter name should fall back to the inviter email")
	}
	if !strings.Contains(html, ">member<") {
		t.Fatal("missing role should fall back to member")
	}
	if !strings.Contains(html, strings.Repeat("a", 160)) {
		t.Fatal("long invitation token was truncated")
	}

	escaped, _ := renderWorkspaceInvitationForTest(t, WorkspaceInvitationData{
		InviterName:    `<script>alert("x")</script>`,
		InviterEmail:   "safe@example.com",
		WorkspaceName:  `<img src=x onerror=alert(1)>`,
		Role:           "member",
		InvitationLink: "https://app.businessos.dev/invite/safe",
		ExpiresAt:      "on September 15, 2026",
	})
	if strings.Contains(escaped, "<script>") || strings.Contains(escaped, "<img src=x") {
		t.Fatal("invitation content must remain HTML escaped")
	}
}
