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

	cmd, err := ParseCommand("add link my-span another-span")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.AddLink, "AddLink command should not be nil")

	handleAddLinkCommand(cmd.AddLink)

	span, exists := telemetry.GetSpans()["my-span"]
	assert.True(t, exists)

	links := span.Links
	assert.Len(t, links, 1, "Span should have one link")
	assert.Equal(t, "another-span", links[0].Name, "Link should point to 'another-span'")
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
