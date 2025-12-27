package streaming

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ArtifactDetector detects and extracts artifacts from streaming output
type ArtifactDetector struct {
	buffer         strings.Builder
	inArtifact     bool
	artifactBuffer strings.Builder
	startPattern   *regexp.Regexp
}

// NewArtifactDetector creates a new artifact detector
func NewArtifactDetector() *ArtifactDetector {
	return &ArtifactDetector{
		startPattern: regexp.MustCompile("```artifact\\s*\\n?"),
	}
}

// ProcessChunk processes a chunk of streaming output and returns events
func (d *ArtifactDetector) ProcessChunk(chunk string) []StreamEvent {
	var events []StreamEvent

	d.buffer.WriteString(chunk)
	content := d.buffer.String()

	if !d.inArtifact {
		events = append(events, d.processNormalContent(content)...)
	} else {
		events = append(events, d.processArtifactContent(chunk)...)
	}

	return events
}

// processNormalContent handles content when not inside an artifact
func (d *ArtifactDetector) processNormalContent(content string) []StreamEvent {
	var events []StreamEvent

	startMatch := d.startPattern.FindStringIndex(content)

	if startMatch == nil {
		if len(content) > 15 {
			safeContent := content[:len(content)-15]
			events = append(events, StreamEvent{Type: EventTypeToken, Content: safeContent})
			d.buffer.Reset()
			d.buffer.WriteString(content[len(content)-15:])
		}
		return events
	}

	if startMatch[0] > 0 {
		events = append(events, StreamEvent{Type: EventTypeToken, Content: content[:startMatch[0]]})
	}

	events = append(events, StreamEvent{Type: EventTypeArtifactStart})

	d.inArtifact = true
	d.artifactBuffer.Reset()

	afterMarker := content[startMatch[1]:]
	d.buffer.Reset()
	d.artifactBuffer.WriteString(afterMarker)

	events = append(events, d.checkArtifactComplete()...)

	return events
}

// processArtifactContent handles content when inside an artifact
func (d *ArtifactDetector) processArtifactContent(chunk string) []StreamEvent {
	d.artifactBuffer.WriteString(chunk)
	return d.checkArtifactComplete()
}

// checkArtifactComplete checks if the artifact has ended and parses it
func (d *ArtifactDetector) checkArtifactComplete() []StreamEvent {
	var events []StreamEvent
	content := d.artifactBuffer.String()

	closingIdx := strings.LastIndex(content, "```")
	if closingIdx == -1 {
		return events
	}

	afterClosing := strings.TrimSpace(content[closingIdx+3:])
	if len(afterClosing) > 20 {
		return events
	}

	artifactJSON := strings.TrimSpace(content[:closingIdx])

	var artifact Artifact
	if err := json.Unmarshal([]byte(artifactJSON), &artifact); err != nil {
		events = append(events, StreamEvent{
			Type:    EventTypeArtifactError,
			Content: "Failed to parse artifact: " + err.Error(),
		})
		d.inArtifact = false
		d.buffer.Reset()
		d.buffer.WriteString(afterClosing)
		return events
	}

	events = append(events, StreamEvent{
		Type: EventTypeArtifactComplete,
		Data: artifact,
	})

	d.inArtifact = false
	d.artifactBuffer.Reset()
	d.buffer.Reset()
	d.buffer.WriteString(afterClosing)

	return events
}

// Flush returns any remaining buffered content
func (d *ArtifactDetector) Flush() []StreamEvent {
	var events []StreamEvent

	if d.inArtifact {
		events = append(events, StreamEvent{
			Type:    EventTypeArtifactError,
			Content: "Artifact was not properly closed",
		})
		events = append(events, StreamEvent{
			Type:    EventTypeToken,
			Content: d.artifactBuffer.String(),
		})
	} else if d.buffer.Len() > 0 {
		events = append(events, StreamEvent{
			Type:    EventTypeToken,
			Content: d.buffer.String(),
		})
	}

	d.buffer.Reset()
	d.artifactBuffer.Reset()
	d.inArtifact = false

	return events
}

// Reset resets the detector state
func (d *ArtifactDetector) Reset() {
	d.buffer.Reset()
	d.artifactBuffer.Reset()
	d.inArtifact = false
}

// IsInArtifact returns whether we're currently inside an artifact
func (d *ArtifactDetector) IsInArtifact() bool {
	return d.inArtifact
}
