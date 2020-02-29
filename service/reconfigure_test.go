package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/picostack/picobot/service/task"
)

func Test_diffTargets(t *testing.T) {
	type args struct {
		oldTargets []task.Target
		newTargets []task.Target
	}
	tests := []struct {
		args          args
		wantAdditions []task.Target
		wantRemovals  []task.Target
	}{
		{
			args{
				oldTargets: []task.Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
				newTargets: []task.Target{
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
				oldTargets: []task.Target{},
				newTargets: []task.Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
			},
			[]task.Target{
				{Name: "one"},
				{Name: "two"},
				{Name: "three"},
			},
			nil,
		},
		{
			args{
				oldTargets: []task.Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
				newTargets: []task.Target{},
			},
			nil,
			[]task.Target{
				{Name: "one"},
				{Name: "two"},
				{Name: "three"},
			},
		},
		{
			args{
				oldTargets: []task.Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
				newTargets: []task.Target{
					{Name: "one"},
					{Name: "three"},
				},
			},
			nil,
			[]task.Target{
				{Name: "two"},
			},
		},
		{
			args{
				oldTargets: []task.Target{
					{Name: "one"},
					{Name: "three"},
				},
				newTargets: []task.Target{
					{Name: "one"},
					{Name: "two"},
					{Name: "three"},
				},
			},
			[]task.Target{
				{Name: "two"},
			},
			nil,
		},
		{
			args{
				oldTargets: []task.Target{
					{Name: "one"},
					{Name: "two", RepoURL: "123"},
					{Name: "three"},
				},
				newTargets: []task.Target{
					{Name: "one"},
					{Name: "two", RepoURL: "312"},
					{Name: "three"},
				},
			},
			[]task.Target{
				{Name: "two", RepoURL: "312"},
			},
			nil,
		},
	}
	for ii, tt := range tests {
		t.Run(fmt.Sprint(ii), func(t *testing.T) {
			gotAdditions, gotRemovals := diffTargets(tt.args.oldTargets, tt.args.newTargets)
			assert.Equal(t, tt.wantAdditions, gotAdditions, "additions mismatch")
			assert.Equal(t, tt.wantRemovals, gotRemovals, "removals mismatch")
		})
	}
}
