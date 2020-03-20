package task

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DiffTargets(t *testing.T) {
	type args struct {
		oldTargets []Target
		newTargets []Target
	}
	tests := []struct {
		args          args
		wantAdditions []Target
		wantRemovals  []Target
	}{
		{
			args{
				oldTargets: []Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
				newTargets: []Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
			},
			nil,
			nil,
		},
		{
			args{
				oldTargets: []Target{},
				newTargets: []Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
			},
			[]Target{
				{Name: "one"},
				{Name: "two"},
				{Name: "three"},
			},
			nil,
		},
		{
			args{
				oldTargets: []Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
				newTargets: []Target{},
			},
			nil,
			[]Target{
				{Name: "one"},
				{Name: "two"},
				{Name: "three"},
			},
		},
		{
			args{
				oldTargets: []Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
				newTargets: []Target{
					{Name: "one"},
					{Name: "three"},
				},
			},
			nil,
			[]Target{
				{Name: "two"},
			},
		},
		{
			args{
				oldTargets: []Target{
					{Name: "one"},
					{Name: "three"},
				},
				newTargets: []Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
			},
			[]Target{
				{Name: "two"},
			},
			nil,
		},
		{
			args{
				oldTargets: []Target{
					{Name: "one"},
					{Name: "two", RepoURL: "123"},
					{Name: "three"},
				},
				newTargets: []Target{
					{Name: "one"},
					{Name: "two", RepoURL: "312"},
					{Name: "three"},
				},
			},
			[]Target{
				{Name: "two", RepoURL: "312"},
			},
			nil,
		},
	}
	for ii, tt := range tests {
		t.Run(fmt.Sprint(ii), func(t *testing.T) {
			gotAdditions, gotRemovals := DiffTargets(tt.args.oldTargets, tt.args.newTargets)
			assert.Equal(t, tt.wantAdditions, gotAdditions, "additions mismatch")
			assert.Equal(t, tt.wantRemovals, gotRemovals, "removals mismatch")
		})
	}
}
