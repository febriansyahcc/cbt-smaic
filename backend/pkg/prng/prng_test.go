package prng

import (
	"testing"

	"cbt-backend/internal/domain"

	"github.com/google/uuid"
)

func TestSeededRandomizerConsistency(t *testing.T) {
	student1 := uuid.New().String()
	student2 := uuid.New().String()
	scheduleID := uuid.New().String()

	seed1 := GenerateSeed(student1, scheduleID)
	seed1Again := GenerateSeed(student1, scheduleID)
	seed2 := GenerateSeed(student2, scheduleID)

	if seed1 != seed1Again {
		t.Fatalf("Expected deterministic seed for student1, got %d and %d", seed1, seed1Again)
	}

	if seed1 == seed2 {
		t.Fatalf("Expected different seeds for student1 and student2")
	}

	questions := []domain.Question{
		{ID: uuid.New(), QuestionNumber: 1},
		{ID: uuid.New(), QuestionNumber: 2},
		{ID: uuid.New(), QuestionNumber: 3},
		{ID: uuid.New(), QuestionNumber: 4},
		{ID: uuid.New(), QuestionNumber: 5},
	}

	shuffled1 := ShuffleQuestions(questions, seed1)
	shuffled1Again := ShuffleQuestions(questions, seed1Again)

	for i := range shuffled1 {
		if shuffled1[i].ID != shuffled1Again[i].ID {
			t.Errorf("Mismatch on index %d across identical seed shuffles", i)
		}
	}

	shuffled2 := ShuffleQuestions(questions, seed2)
	// Check that shuffled2 is different in at least one index from shuffled1
	anyDiff := false
	for i := range questions {
		if shuffled1[i].ID != shuffled2[i].ID {
			anyDiff = true
			break
		}
	}
	if !anyDiff {
		t.Logf("Warning: Both shuffles resulted in identical order (possible for small sets)")
	}
}
