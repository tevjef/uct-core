package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"uct/common"
)

func TestFo(t *testing.T) {
	assert.Equal(t, " ", common.TrimAll(" "))
}
