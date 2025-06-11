package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otelgen/telemetry"
)

func TestHandleAddLink_OK(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{})

	telemetry.CreateTrace("another-trace")
	telemetry.AddSpanToTrace("another-trace", "another-span", map[string]string{})

	cmd, err := ParseCommand("add link my-span another-span attributes key=value")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.AddLink, "AddLink command should not be nil")

	handleAddLinkCommand(cmd.AddLink)

	span, exists := telemetry.GetSpans()["my-span"]
	assert.True(t, exists)

	links := span.Links
	assert.Len(t, links, 1, "Span should have one link")
	assert.Equal(t, "another-span", links[0].TargetSpan.Name, "Link should point to 'another-span'")
	assert.Equal(t, "value", links[0].Attributes["key"], "Link should have attribute 'key' with value 'value'")
}

func TestHandleAddLink_NonExistingSpan(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{})

	telemetry.CreateTrace("another-trace")
	telemetry.AddSpanToTrace("another-trace", "another-span", map[string]string{})

	t.Run("from span does not exist", func(t *testing.T) {
		cmd, err := ParseCommand("add link non-existing-span another-span")
		assert.Nil(t, err, "ParseCommand should not return an error")
		assert.NotNil(t, cmd.AddLink, "AddLink command should not be nil")

		output := captureOutput(func() {
			handleAddLinkCommand(cmd.AddLink)
		})

		assert.Equal(t, "Error validating add link command: span 'non-existing-span' does not exist\n", output)
	})

	t.Run("to span does not exist", func(t *testing.T) {
		cmd, err := ParseCommand("add link my-span non-existing-span")
		assert.Nil(t, err, "ParseCommand should not return an error")
		assert.NotNil(t, cmd.AddLink, "AddLink command should not be nil")

		output := captureOutput(func() {
			handleAddLinkCommand(cmd.AddLink)
		})

		assert.Equal(t, "Error validating add link command: span 'non-existing-span' does not exist\n", output)
	})
}

func TestHandleAddEvent_OK(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{})
	telemetry.CreateEvent("my-event", map[string]string{"key": "value"})

	cmd, err := ParseCommand("add event my-span my-event")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.AddEvent, "AddEvent command should not be nil")

	handleAddEventCommand(cmd.AddEvent)

	span, exists := telemetry.GetSpans()["my-span"]
	assert.True(t, exists)

	events := span.Events
	assert.Len(t, events, 1, "Span should have one event")
	assert.Equal(t, "my-event", events[0].Name, "Event should be 'my-event'")
	assert.Equal(t, "value", events[0].Attributes["key"], "Event should have attribute 'key' with value 'value'")
}

func TestHandleAddEvent_NonExistingSpan(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{})
	telemetry.CreateEvent("my-event", map[string]string{"key": "value"})

	cmd, err := ParseCommand("add event non-existing-span my-event")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.AddEvent, "AddEvent command should not be nil")

	output := captureOutput(func() {
		handleAddEventCommand(cmd.AddEvent)
	})

	assert.Equal(t, "Error validating add event command: span 'non-existing-span' does not exist\n", output)
}

func TestHandleAddEvent_NonExistingEvent(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{})
	telemetry.CreateEvent("my-event", map[string]string{"key": "value"})

	cmd, err := ParseCommand("add event my-span non-existing-event")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.AddEvent, "AddEvent command should not be nil")

	output := captureOutput(func() {
		handleAddEventCommand(cmd.AddEvent)
	})

	assert.Equal(t, "Error validating add event command: event 'non-existing-event' does not exist\n", output)
}
