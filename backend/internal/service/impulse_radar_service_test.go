package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImpulseRadarService(t *testing.T) {
	svc := NewImpulseRadarService(nil)
	assert.NotNil(t, svc)
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.True(t, contains(slice, "a"))
	assert.True(t, contains(slice, "b"))
	assert.True(t, contains(slice, "c"))
	assert.False(t, contains(slice, "d"))
}
