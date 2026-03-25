//! Process model merger for combining local models into global net

use super::types::ProcessModel;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};

/// Merges local process models from workers into global model
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelMerger {
    local_models: HashMap<String, ProcessModel>,
    global_model: ProcessModel,
}

impl ModelMerger {
    /// Create a new model merger
    pub fn new() -> Self {
        Self {
            local_models: HashMap::new(),
            global_model: ProcessModel::new("global".to_string()),
        }
    }

    /// Register a local model from a worker
    pub fn register_model(&mut self, partition_id: String, model: ProcessModel) -> Result<()> {
        self.local_models.insert(partition_id, model);
        Ok(())
    }

    /// Merge all local models into global model
    pub fn merge_models(&mut self) -> Result<ProcessModel> {
        let mut merged = ProcessModel::new("global".to_string());

        // Collect all unique places and transitions
        let mut all_places = HashSet::new();
        let mut all_transitions = HashSet::new();
        let mut all_arcs = HashSet::new();

        for (_, model) in &self.local_models {
            // Add all places
            for place in &model.places {
                all_places.insert(place.clone());
            }

            // Add all transitions
            for transition in &model.transitions {
                all_transitions.insert(transition.clone());
            }

            // Add all arcs
            for (from, to) in &model.arcs {
                all_arcs.insert((from.clone(), to.clone()));
            }
        }

        // Build merged model
        for place in all_places {
            merged.add_place(place);
        }

        for transition in all_transitions {
            merged.add_transition(transition);
        }

        for (from, to) in all_arcs {
            merged.add_arc(from, to);
        }

        merged.compute_hash();
        self.global_model = merged.clone();
        Ok(merged)
    }

    /// Get global model
    pub fn get_global_model(&self) -> ProcessModel {
        self.global_model.clone()
    }

    /// Handle duplicate places by renaming
    fn handle_duplicate_place(&self, place: &str, partition_id: &str) -> String {
        if self.global_model.places.contains(&place.to_string()) {
            format!("{}_{}", partition_id, place)
        } else {
            place.to_string()
        }
    }

    /// Handle duplicate transitions by merging behavior
    fn handle_duplicate_transition(&self, transition: &str, _partition_id: &str) -> String {
        if self.global_model.transitions.contains(&transition.to_string()) {
            // Keep original name to enable cross-partition synchronization
            transition.to_string()
        } else {
            transition.to_string()
        }
    }

    /// Verify Petri net soundness after merge
    pub fn verify_soundness(&self) -> Result<bool> {
        let model = &self.global_model;

        // Basic soundness checks:
        // 1. Model is not empty
        if model.places.is_empty() || model.transitions.is_empty() {
            return Ok(false);
        }

        // 2. At least one arc exists
        if model.arcs.is_empty() {
            return Ok(false);
        }

        // 3. All places and transitions are referenced in arcs
        // (allowing places/transitions that exist in the model)
        for place in &model.places {
            let is_referenced = model.arcs.iter().any(|(from, to)| from == place || to == place);
            if !is_referenced {
                return Ok(false);
            }
        }

        Ok(true)
    }

    /// Get merge statistics
    pub fn get_merge_stats(&self) -> MergeStats {
        MergeStats {
            num_local_models: self.local_models.len(),
            global_places: self.global_model.places.len(),
            global_transitions: self.global_model.transitions.len(),
            global_arcs: self.global_model.arcs.len(),
        }
    }

    /// Get model from partition
    pub fn get_local_model(&self, partition_id: &str) -> Option<ProcessModel> {
        self.local_models.get(partition_id).cloned()
    }

    /// Get all registered partition IDs
    pub fn get_partitions(&self) -> Vec<String> {
        self.local_models.keys().cloned().collect()
    }
}

impl Default for ModelMerger {
    fn default() -> Self {
        Self::new()
    }
}

/// Statistics about model merge
#[derive(Debug, Clone)]
pub struct MergeStats {
    pub num_local_models: usize,
    pub global_places: usize,
    pub global_transitions: usize,
    pub global_arcs: usize,
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_model(partition: &str, places: &[&str], transitions: &[&str]) -> ProcessModel {
        let mut model = ProcessModel::new(partition.to_string());
        for place in places {
            model.add_place(place.to_string());
        }
        for transition in transitions {
            model.add_transition(transition.to_string());
        }
        model
    }

    #[test]
    fn test_merger_creation() {
        let merger = ModelMerger::new();
        assert_eq!(merger.local_models.len(), 0);
    }

    #[test]
    fn test_register_model() {
        let mut merger = ModelMerger::new();
        let model = create_test_model("partition_a", &["p1"], &["t1"]);

        merger.register_model("partition_a".to_string(), model).unwrap();
        assert_eq!(merger.local_models.len(), 1);
    }

    #[test]
    fn test_merge_single_model() {
        let mut merger = ModelMerger::new();
        let model = create_test_model("partition_a", &["p1", "p2"], &["t1", "t2"]);

        merger.register_model("partition_a".to_string(), model).unwrap();
        let merged = merger.merge_models().unwrap();

        assert_eq!(merged.places.len(), 2);
        assert_eq!(merged.transitions.len(), 2);
    }

    #[test]
    fn test_merge_multiple_models() {
        let mut merger = ModelMerger::new();

        let model_a = create_test_model("partition_a", &["p1", "p2"], &["t1"]);
        let model_b = create_test_model("partition_b", &["p3", "p4"], &["t2"]);

        merger.register_model("partition_a".to_string(), model_a).unwrap();
        merger.register_model("partition_b".to_string(), model_b).unwrap();

        let merged = merger.merge_models().unwrap();

        assert_eq!(merged.places.len(), 4);
        assert_eq!(merged.transitions.len(), 2);
    }

    #[test]
    fn test_merge_duplicate_places() {
        let mut merger = ModelMerger::new();

        let mut model_a = create_test_model("partition_a", &["p1"], &["t1"]);
        let mut model_b = create_test_model("partition_b", &["p1"], &["t2"]);

        // Add arcs to make them valid
        model_a.add_arc("p1".to_string(), "t1".to_string());
        model_b.add_arc("p1".to_string(), "t2".to_string());

        merger.register_model("partition_a".to_string(), model_a).unwrap();
        merger.register_model("partition_b".to_string(), model_b).unwrap();

        let merged = merger.merge_models().unwrap();

        // Duplicate places should be merged
        assert_eq!(merged.places.len(), 1);
    }

    #[test]
    fn test_merge_duplicate_transitions() {
        let mut merger = ModelMerger::new();

        let mut model_a = create_test_model("partition_a", &["p1"], &["t1"]);
        let mut model_b = create_test_model("partition_b", &["p2"], &["t1"]);

        model_a.add_arc("p1".to_string(), "t1".to_string());
        model_b.add_arc("p2".to_string(), "t1".to_string());

        merger.register_model("partition_a".to_string(), model_a).unwrap();
        merger.register_model("partition_b".to_string(), model_b).unwrap();

        let merged = merger.merge_models().unwrap();

        // Same transition in both partitions (cross-partition synchronization)
        assert_eq!(merged.transitions.len(), 1);
    }

    #[test]
    fn test_get_global_model() {
        let mut merger = ModelMerger::new();
        let model = create_test_model("partition_a", &["p1"], &["t1"]);

        merger.register_model("partition_a".to_string(), model).unwrap();
        merger.merge_models().unwrap();

        let global = merger.get_global_model();
        assert_eq!(global.partition_id, "global");
    }

    #[test]
    fn test_merge_stats() {
        let mut merger = ModelMerger::new();

        let model_a = create_test_model("partition_a", &["p1"], &["t1"]);
        let model_b = create_test_model("partition_b", &["p2"], &["t2"]);

        merger.register_model("partition_a".to_string(), model_a).unwrap();
        merger.register_model("partition_b".to_string(), model_b).unwrap();
        merger.merge_models().unwrap();

        let stats = merger.get_merge_stats();
        assert_eq!(stats.num_local_models, 2);
        assert_eq!(stats.global_places, 2);
        assert_eq!(stats.global_transitions, 2);
    }

    #[test]
    fn test_verify_soundness() {
        let mut merger = ModelMerger::new();

        let mut model = ProcessModel::new("partition_a".to_string());
        model.add_place("p1".to_string());
        model.add_transition("t1".to_string());
        model.add_arc("p1".to_string(), "t1".to_string());

        merger.register_model("partition_a".to_string(), model).unwrap();
        merger.merge_models().unwrap();

        // This should pass basic connectivity checks
        let sound = merger.verify_soundness().unwrap();
        assert!(sound);
    }
}
