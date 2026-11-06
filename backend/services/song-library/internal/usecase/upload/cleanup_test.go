package upload

import (
	"errors"
	"reflect"
	"testing"
)

func TestCleanupContinuesAfterOneRemovalFails(t *testing.T) {
	var calls []string
	repo := uploadRepository{calls: &calls}
	files := fakeFiles{
		calls: &calls,
		err:   errors.New("storage offline"),
	}
	if err := New(repo, files).CleanupExpiredAttempts(t.Context()); err == nil {
		t.Fatal("lost failed cleanup")
	}
	want := []string{"remove expired", "remove late"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls=%v want=%v", calls, want)
	}
	calls = nil
	files.err = nil
	if err := New(repo, files).CleanupExpiredAttempts(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("retry did not revisit expired keys: %v", calls)
	}
}
