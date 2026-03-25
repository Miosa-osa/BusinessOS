// Code generated from semconv/model/business_os/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 11

package semconv

import "go.opentelemetry.io/otel/attribute"

// BusinessOS compliance, audit and integration attributes (iter11).
// Note: business_os.* prefix is distinct from bos.* prefix used in bos_attributes.go.

const (
	// BusinessOsComplianceFrameworkKey is the OTel attribute key for business_os.compliance.framework.
	// Compliance framework name (e.g. SOC2, HIPAA, GDPR) for BusinessOS audit spans.
	BusinessOsComplianceFrameworkKey = attribute.Key("business_os.compliance.framework")
	// BusinessOsAuditEventTypeKey is the OTel attribute key for business_os.audit.event_type.
	// Type of audit event emitted by BusinessOS (e.g. data_access, config_change).
	BusinessOsAuditEventTypeKey = attribute.Key("business_os.audit.event_type")
	// BusinessOsIntegrationTypeKey is the OTel attribute key for business_os.integration.type.
	// Type of external integration active in BusinessOS (e.g. google, hubspot, notion).
	BusinessOsIntegrationTypeKey = attribute.Key("business_os.integration.type")
)

// BusinessOsComplianceFramework returns an attribute KeyValue for business_os.compliance.framework.
func BusinessOsComplianceFramework(val string) attribute.KeyValue {
	return BusinessOsComplianceFrameworkKey.String(val)
}

// BusinessOsAuditEventType returns an attribute KeyValue for business_os.audit.event_type.
func BusinessOsAuditEventType(val string) attribute.KeyValue {
	return BusinessOsAuditEventTypeKey.String(val)
}

// BusinessOsIntegrationType returns an attribute KeyValue for business_os.integration.type.
func BusinessOsIntegrationType(val string) attribute.KeyValue {
	return BusinessOsIntegrationTypeKey.String(val)
}
