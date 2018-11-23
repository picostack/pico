package service

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Southclaws/wadsworth/service/task"
)

func Test_diffTargets(t *testing.T) {
	type args struct {
		newTargets []task.Target
		oldTargets []task.Target
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
			gotAdditions, gotRemovals := diffTargets(tt.args.newTargets, tt.args.oldTargets)
			if !reflect.DeepEqual(gotAdditions, tt.wantAdditions) {
				t.Errorf("diffTargets() gotAdditions = %v, want %v", gotAdditions, tt.wantAdditions)
			}
			if !reflect.DeepEqual(gotRemovals, tt.wantRemovals) {
				t.Errorf("diffTargets() gotRemovals = %v, want %v", gotRemovals, tt.wantRemovals)
			}
		})
	}
}
