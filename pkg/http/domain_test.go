package http

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPParseDomain(t *testing.T) {
	full, err := ParseDomain("inference-enclave.tinfoil.sh")
	assert.Nil(t, err)
	assert.Equal(t, "inference-enclave.tinfoil.sh", full)

	first, err := ParseDomain("*.alpha.tinfoil.sh")
	assert.Nil(t, err)
	firstPrefix := strings.TrimSuffix(first, ".alpha.tinfoil.sh")
	assert.True(t, strings.HasSuffix(first, ".alpha.tinfoil.sh"))

	second, err := ParseDomain("*.bravo.tinfoil.sh")
	assert.Nil(t, err)
	secondPrefix := strings.TrimSuffix(second, ".bravo.tinfoil.sh")
	assert.True(t, strings.HasSuffix(second, ".bravo.tinfoil.sh"))

	assert.NotEqual(t, firstPrefix, secondPrefix)
}
