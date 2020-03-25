package task

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareTargetExecution(t *testing.T) {
	c, err := prepare(".", map[string]string{
		"VAR_1": "one",
		"VAR_2": "two",
		"VAR_3": "three",
		"VAR_4": "four",
	}, []string{"docker-compose", "up", "-d"}, false)
	assert.NoError(t, err)

	assert.Equal(t, []string{"docker-compose", "up", "-d"}, c.Args)
	want := []string{
		"VAR_1=one",
		"VAR_2=two",
		"VAR_3=three",
		"VAR_4=four",
	}
	got := c.Env
	sort.Strings(want)
	sort.Strings(got)
	assert.Equal(t, want, got)
	assert.Equal(t, ".", c.Dir)
}
