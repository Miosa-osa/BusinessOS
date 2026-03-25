//! bos-core — BusinessOS data layer core.
//!
//! Provides ODCS workspace operations, schema validation/conversion,
//! MADR decision records, knowledge base management, RDF triple store,
//! ontology bridge capabilities, and supervision tree fault tolerance.

pub mod decisions;
pub mod export;
pub mod gateway;
pub mod knowledge;
pub mod ontology;
pub mod rdf;
pub mod schema;
pub mod workspace;
pub mod supervision;
pub mod distributed;
pub mod streaming;

pub use gateway::{
    BusinessOSGateway, GatewayConfig, GatewayStatus, DiscoverRequest, DiscoverResponse,
    ConformanceRequest, ConformanceResponse, StatisticsRequest, StatisticsResponse,
};
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
pub use ontology::serve::{ServeConfig, serve as serve_ontology};
pub use rdf::store::{TripleStore, Triple};
pub use supervision::{SupervisorConfig, SupervisorHandle, Worker, WorkerConfig, WorkerHandle, WorkerState};
pub use distributed::{DistributedPM, RaftCoordinator, DistributedWorker, ConsensusProtocol, ModelMerger, FaultRecovery};
pub use streaming::{
    StreamingCoordinator, StreamingSessionHandle, StreamEvent, StreamEventType,
    ProgressEvent, MetricsEvent, PartialResultEvent, ErrorEvent,
};

// Process Mining (pm4py-rust integration)
pub mod process;
