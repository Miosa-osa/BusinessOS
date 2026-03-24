//! bos-core — BusinessOS data layer core.
//!
//! Provides ODCS workspace operations, schema validation/conversion,
//! MADR decision records, knowledge base management, RDF triple store,
//! and ontology bridge capabilities.

pub mod decisions;
pub mod export;
pub mod knowledge;
pub mod ontology;
pub mod rdf;
pub mod schema;
pub mod workspace;

pub use workspace::{WorkspaceGenerator, WorkspaceInitOptions, WorkspaceValidationResult};
pub use schema::{SchemaConverter, SchemaValidationResult, FormatHint};
pub use decisions::{DecisionGenerator, DecisionRecord, DecisionIndex};
pub use knowledge::{KnowledgeBase, KnowledgeArticle, KnowledgeIndex};
pub use ontology::construct::ConstructGenerator;
pub use ontology::mapping::{MappingConfig, TableMapping, PropertyMapping};
pub use ontology::bridge::BusinessOSBridge;
pub use ontology::execute::{QueryExecutor, ExecutionResult};
pub use ontology::infer::{ConfidenceLevel, InferConfig, InferResult, OntologyInferrer};
pub use ontology::select::{SemanticSearch, SelectResult};
pub use rdf::store::{TripleStore, Triple};
