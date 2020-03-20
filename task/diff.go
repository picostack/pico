package task

import "reflect"

// DiffTargets returns just the additions (also changes) and removals between
// the specified old targets and new targets
func DiffTargets(oldTargets, newTargets []Target) (additions, removals []Target) {
	for _, newTarget := range newTargets {
		var exists bool
		for _, oldTarget := range oldTargets {
			if oldTarget.Name == newTarget.Name {
				exists = true
				break
			}
		}
		if !exists {
			additions = append(additions, newTarget)
		}
	}
	for _, oldTarget := range oldTargets {
		var exists bool
		var newTarget Target
		for _, newTarget = range newTargets {
			if newTarget.Name == oldTarget.Name {
				exists = true
				break
			}
		}
		if !exists {
			removals = append(removals, oldTarget)
		} else if !reflect.DeepEqual(oldTarget, newTarget) {
			additions = append(additions, newTarget)
		}
	}
	return
}
