package redis

import (
	"testing"
)

// tested: string[] -> expected : []BulkString

func commandArrayCreate(strVals ...string) []string {
	return strVals
}

func TestExecBehavior(t *testing.T) {
	tClient := newStringCommandTestClient()
	_ = tClient

}
