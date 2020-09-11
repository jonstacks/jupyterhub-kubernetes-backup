package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanUserName(t *testing.T) {
	assert.Equal(t, "bo.zhou", getUserNameFromPVCName("claim-bo-2ezhou"))
	assert.Equal(t, "anthony.shipman", getUserNameFromPVCName("claim-anthony-20shipman"))
}