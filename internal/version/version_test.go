// SPDX-License-Identifier: Apache-2.0

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion_Default(t *testing.T) {
	original := version
	version = ""
	defer func() { version = original }()

	assert.Equal(t, "0.0.0-unknown", Version())
}

func TestVersion_WithVPrefix(t *testing.T) {
	original := version
	version = "v0.1.0"
	defer func() { version = original }()

	assert.Equal(t, "0.1.0", Version())
}

func TestVersion_WithoutPrefix(t *testing.T) {
	original := version
	version = "0.2.0"
	defer func() { version = original }()

	assert.Equal(t, "0.2.0", Version())
}
