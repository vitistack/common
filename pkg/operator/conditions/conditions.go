package conditions

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get returns the condition with the given type, if present.
func Get(conds []metav1.Condition, t string) (metav1.Condition, bool) {
	for i := range conds {
		if conds[i].Type == t {
			return conds[i], true
		}
	}
	return metav1.Condition{}, false
}

// Has returns true if a condition of the given type exists.
func Has(conds []metav1.Condition, t string) bool {
	_, ok := Get(conds, t)
	return ok
}

// SetOrUpdateCondition inserts or updates a condition in-place.
// - Preserves LastTransitionTime when Status is unchanged.
// - Updates LastTransitionTime to now when Status changes.
// - Replaces Reason, Message, and ObservedGeneration with the provided values.
// Returns true when the slice was modified.
func SetOrUpdateCondition(conds *[]metav1.Condition, cond *metav1.Condition) bool {
	if conds == nil || cond == nil {
		return false
	}
	now := metav1.NewTime(time.Now())
	s := *conds
	v := *cond // work on a local copy
	for i := range s {
		if s[i].Type == v.Type {
			// Existing condition: check transition
			if s[i].Status != v.Status {
				v.LastTransitionTime = now
			} else {
				// Preserve prior transition time
				v.LastTransitionTime = s[i].LastTransitionTime
			}
			// Always carry ObservedGeneration from input when non-zero; otherwise keep old
			if v.ObservedGeneration == 0 {
				v.ObservedGeneration = s[i].ObservedGeneration
			}
			// Replace entry
			s[i] = v
			*conds = s
			return true
		}
	}
	// Not found: set transition time and append
	if v.LastTransitionTime.IsZero() {
		v.LastTransitionTime = now
	}
	s = append(s, v)
	*conds = s
	return true
}

// Remove removes a condition by type. Returns true when removed.
func Remove(conds *[]metav1.Condition, t string) bool {
	if conds == nil {
		return false
	}
	s := *conds
	for i := range s {
		if s[i].Type == t {
			s = append(s[:i], s[i+1:]...)
			*conds = s
			return true
		}
	}
	return false
}

// New creates a new condition with transition time set to now.
func New(t string, status metav1.ConditionStatus, reason, message string, observedGen int64) metav1.Condition {
	return metav1.Condition{
		Type:               t,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: observedGen,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
}
