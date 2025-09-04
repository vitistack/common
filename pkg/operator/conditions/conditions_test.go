package conditions

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetOrUpdateCondition(t *testing.T) {
	var conds []metav1.Condition
	c0 := New("Ready", metav1.ConditionFalse, "Init", "starting", 1)
	updated := SetOrUpdateCondition(&conds, &c0)
	if !updated || len(conds) != 1 {
		t.Fatalf("expected add")
	}
	first := conds[0]
	// Update without status change should preserve transition time
	c1 := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             "StillInit",
		Message:            "waiting",
		ObservedGeneration: 2,
	}
	updated = SetOrUpdateCondition(&conds, &c1)
	if !updated || len(conds) != 1 {
		t.Fatalf("expected update")
	}
	if !conds[0].LastTransitionTime.Equal(&first.LastTransitionTime) {
		t.Fatalf("transition time should be preserved when status unchanged")
	}
	// Change status should bump transition time
	c2 := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "Ok",
		Message:            "ready",
		ObservedGeneration: 3,
	}
	updated = SetOrUpdateCondition(&conds, &c2)
	if !updated || len(conds) != 1 {
		t.Fatalf("expected update 2")
	}
	if conds[0].LastTransitionTime.Equal(&first.LastTransitionTime) {
		t.Fatalf("transition time should change when status changed")
	}
}
