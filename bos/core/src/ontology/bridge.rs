//! BusinessOS Bridge -- connects ODCS workspace to RDF/OWL ontology.
//!
//! Provides high-level operations for converting workspace definitions
//! into ontology-aligned RDF representations.

use crate::ontology::mapping::MappingConfig;
use crate::ontology::mapping::ResolvedPrefixes;

/// Bridge between ODCS workspace definitions and the RDF/OWL ontology layer.
pub struct BusinessOSBridge {
    config: MappingConfig,
    prefixes: ResolvedPrefixes,
}

impl BusinessOSBridge {
    /// Create a new bridge from a mapping configuration.
    pub fn new(config: MappingConfig) -> Self {
        let prefixes = ResolvedPrefixes::from_config(&config);
        Self { config, prefixes }
    }

    /// Resolve a prefixed URI to a full IRI.
    pub fn resolve(&self, prefixed: &str) -> String {
        self.prefixes.resolve(prefixed)
    }

    /// Return a reference to the underlying mapping configuration.
    pub fn config(&self) -> &MappingConfig {
        &self.config
    }

    /// Return the SPARQL PREFIX block for this configuration.
    pub fn sparql_prefixes(&self) -> String {
        self.prefixes.to_sparql_prefixes()
    }

    /// Return the list of mapped table names.
    pub fn mapped_tables(&self) -> Vec<&str> {
        self.config.mappings.iter().map(|m| m.table.as_str()).collect()
    }
}
