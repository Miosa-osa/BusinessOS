//! Output formatting module — JSON, table, and graph formatting.

use serde_json::{json, Value};
use std::fmt::Write as FmtWrite;

/// Output format enumeration.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum OutputFormat {
    Json,
    PrettyJson,
    Table,
    Csv,
    PlainText,
    GraphViz,
}

impl std::str::FromStr for OutputFormat {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "json" => Ok(OutputFormat::Json),
            "pretty" | "pretty-json" => Ok(OutputFormat::PrettyJson),
            "table" => Ok(OutputFormat::Table),
            "csv" => Ok(OutputFormat::Csv),
            "text" | "plain" => Ok(OutputFormat::PlainText),
            "graphviz" | "dot" => Ok(OutputFormat::GraphViz),
            _ => Err(format!("Unknown format: {}", s)),
        }
    }
}

/// Result formatter for different output formats.
pub struct ResultFormatter;

impl ResultFormatter {
    /// Format result based on output format specification.
    pub fn format(result: &Value, format: OutputFormat) -> String {
        match format {
            OutputFormat::Json => serde_json::to_string(result).unwrap_or_default(),
            OutputFormat::PrettyJson => {
                serde_json::to_string_pretty(result).unwrap_or_default()
            }
            OutputFormat::Table => Self::format_as_table(result),
            OutputFormat::Csv => Self::format_as_csv(result),
            OutputFormat::PlainText => Self::format_as_plaintext(result),
            OutputFormat::GraphViz => Self::format_as_graphviz(result),
        }
    }

    /// Format as human-readable table.
    fn format_as_table(result: &Value) -> String {
        let mut output = String::new();

        match result {
            Value::Object(map) => {
                for (key, value) in map {
                    match value {
                        Value::Number(n) => {
                            let _ = writeln!(output, "{:.<40} {}", key, n);
                        }
                        Value::String(s) => {
                            let _ = writeln!(output, "{:.<40} {}", key, s);
                        }
                        Value::Bool(b) => {
                            let _ = writeln!(output, "{:.<40} {}", key, b);
                        }
                        Value::Array(arr) => {
                            let _ = writeln!(output, "{:.<40} (array with {} items)", key, arr.len());
                        }
                        _ => {
                            let _ = writeln!(output, "{:.<40} (object)", key);
                        }
                    }
                }
            }
            Value::Array(arr) => {
                let _ = writeln!(output, "┌─ Array ({} items) ─┐", arr.len());
                for (idx, item) in arr.iter().enumerate() {
                    let _ = writeln!(output, "  [{}]: {}", idx, item);
                }
                let _ = writeln!(output, "└──────────────────┘");
            }
            _ => {
                let _ = write!(output, "{}", result);
            }
        }

        output
    }

    /// Format as CSV.
    fn format_as_csv(result: &Value) -> String {
        let mut output = String::new();

        match result {
            Value::Array(arr) => {
                // Extract headers from first object if available
                if let Some(Value::Object(first)) = arr.first() {
                    let headers: Vec<String> = first.keys().cloned().collect();
                    let _ = writeln!(output, "{}", headers.join(","));

                    // Write rows
                    for item in arr {
                        if let Value::Object(obj) = item {
                            let values: Vec<String> = headers
                                .iter()
                                .map(|h| {
                                    obj.get(h)
                                        .map(|v| Self::csv_escape(v.to_string()))
                                        .unwrap_or_default()
                                })
                                .collect();
                            let _ = writeln!(output, "{}", values.join(","));
                        }
                    }
                }
            }
            Value::Object(obj) => {
                let _ = writeln!(output, "key,value");
                for (k, v) in obj {
                    let _ = writeln!(output, "{},{}", Self::csv_escape(k.clone()), Self::csv_escape(v.to_string()));
                }
            }
            _ => {
                let _ = write!(output, "{}", result);
            }
        }

        output
    }

    /// Format as plain text.
    fn format_as_plaintext(result: &Value) -> String {
        match result {
            Value::Object(map) => {
                let mut output = String::new();
                for (key, value) in map {
                    let _ = writeln!(output, "{}: {}", key, Self::value_to_display_text(value));
                }
                output
            }
            Value::Array(arr) => {
                let mut output = String::new();
                for (idx, item) in arr.iter().enumerate() {
                    let _ = writeln!(output, "[{}] {}", idx, Self::value_to_display_text(item));
                }
                output
            }
            _ => result.to_string(),
        }
    }

    /// Format as GraphViz DOT format (for process models).
    fn format_as_graphviz(result: &Value) -> String {
        let mut output = String::from("digraph ProcessModel {\n");
        output.push_str("  rankdir=LR;\n");
        output.push_str("  node [shape=box, style=filled, fillcolor=lightblue];\n");

        if let Value::Object(obj) = result {
            // Extract places and transitions if available
            if let Some(Value::Number(places)) = obj.get("places") {
                let _ = writeln!(output, "  // Places: {}", places);
            }
            if let Some(Value::Number(transitions)) = obj.get("transitions") {
                let _ = writeln!(output, "  // Transitions: {}", transitions);
            }

            // Add a simple node structure
            output.push_str("  place_start [label=\"Start\"];\n");
            output.push_str("  place_end [label=\"End\"];\n");
            output.push_str("  place_start -> place_end;\n");
        }

        output.push_str("}\n");
        output
    }

    /// Escape CSV field values.
    fn csv_escape(value: String) -> String {
        if value.contains(',') || value.contains('"') || value.contains('\n') {
            format!("\"{}\"", value.replace('"', "\"\""))
        } else {
            value
        }
    }

    /// Convert value to display text.
    fn value_to_display_text(value: &Value) -> String {
        match value {
            Value::Number(n) => n.to_string(),
            Value::String(s) => s.clone(),
            Value::Bool(b) => b.to_string(),
            Value::Null => "null".to_string(),
            Value::Array(arr) => {
                format!("[{}]", arr.iter()
                    .map(|v| v.to_string())
                    .collect::<Vec<_>>()
                    .join(", "))
            }
            Value::Object(_) => "[object]".to_string(),
        }
    }
}

/// Pretty-print utility for aligned output.
pub struct PrettyTable {
    headers: Vec<String>,
    rows: Vec<Vec<String>>,
}

impl PrettyTable {
    /// Create a new table.
    pub fn new(headers: Vec<String>) -> Self {
        Self {
            headers,
            rows: vec![],
        }
    }

    /// Add a row to the table.
    pub fn add_row(&mut self, row: Vec<String>) {
        self.rows.push(row);
    }

    /// Render table as aligned text.
    pub fn render(&self) -> String {
        if self.headers.is_empty() {
            return String::new();
        }

        // Calculate column widths
        let mut col_widths = vec![0; self.headers.len()];
        for (i, header) in self.headers.iter().enumerate() {
            col_widths[i] = header.len();
        }
        for row in &self.rows {
            for (i, cell) in row.iter().enumerate() {
                col_widths[i] = col_widths[i].max(cell.len());
            }
        }

        let mut output = String::new();

        // Header separator
        output.push('┌');
        for (i, width) in col_widths.iter().enumerate() {
            output.push_str(&"─".repeat(width + 2));
            if i < col_widths.len() - 1 {
                output.push('┬');
            }
        }
        output.push_str("┐\n");

        // Headers
        output.push('│');
        for (i, header) in self.headers.iter().enumerate() {
            let padding = col_widths[i] - header.len();
            let _ = write!(output, " {} {}", header, " ".repeat(padding));
            output.push('│');
        }
        output.push('\n');

        // Header separator
        output.push('├');
        for (i, width) in col_widths.iter().enumerate() {
            output.push_str(&"─".repeat(width + 2));
            if i < col_widths.len() - 1 {
                output.push('┼');
            }
        }
        output.push_str("┤\n");

        // Rows
        for row in &self.rows {
            output.push('│');
            for (i, cell) in row.iter().enumerate() {
                let padding = col_widths[i] - cell.len();
                let _ = write!(output, " {} {}", cell, " ".repeat(padding));
                output.push('│');
            }
            output.push('\n');
        }

        // Bottom border
        output.push('└');
        for (i, width) in col_widths.iter().enumerate() {
            output.push_str(&"─".repeat(width + 2));
            if i < col_widths.len() - 1 {
                output.push('┴');
            }
        }
        output.push_str("┘\n");

        output
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_json_formatting() {
        let data = json!({"command": "discover", "status": "success"});
        let formatted = ResultFormatter::format(&data, OutputFormat::Json);
        assert!(formatted.contains("discover"));
    }

    #[test]
    fn test_table_formatting() {
        let data = json!({"fitness": 0.85, "precision": 0.78});
        let formatted = ResultFormatter::format(&data, OutputFormat::Table);
        assert!(formatted.contains("fitness"));
        assert!(formatted.contains("0.85"));
    }

    #[test]
    fn test_csv_escaping() {
        let escaped = ResultFormatter::csv_escape("value,with,commas".to_string());
        assert_eq!(escaped, r#""value,with,commas""#);
    }

    #[test]
    fn test_pretty_table() {
        let mut table = PrettyTable::new(vec!["Name".to_string(), "Value".to_string()]);
        table.add_row(vec!["Fitness".to_string(), "0.85".to_string()]);
        table.add_row(vec!["Precision".to_string(), "0.78".to_string()]);

        let rendered = table.render();
        assert!(rendered.contains("Fitness"));
        assert!(rendered.contains("0.85"));
    }
}
