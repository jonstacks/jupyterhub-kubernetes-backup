package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanUserName(t *testing.T) {
	assert.Equal(t, "test.user", GetUserNameFromPVCName("claim-test-2euser"))
	assert.Equal(t, "test.user", GetUserNameFromPVCName("claim-test-20user"))
	assert.Equal(t, "test-user", GetUserNameFromPVCName("claim-test-user-40domain-com"))
}
