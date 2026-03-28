package semconv

const (
	// oxigraph_query is the span name for "oxigraph.query".
	//
	// Oxigraph query — runs SPARQL SELECT, ASK, or CONSTRUCT against the /query endpoint.
	// Kind: client
	// Stability: development
	OxigraphQuerySpan = "oxigraph.query"
	// oxigraph_write is the span name for "oxigraph.write".
	//
	// Oxigraph write — loads Turtle or N-Triples RDF data into Oxigraph via HTTP POST /store.
	// Kind: client
	// Stability: development
	OxigraphWriteSpan = "oxigraph.write"
)