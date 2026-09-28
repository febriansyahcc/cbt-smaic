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

func TestShuffleByGroupKeepsTypesSegregated(t *testing.T) {
	mcIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	esIDs := []uuid.UUID{uuid.New(), uuid.New()}

	questions := []domain.Question{
		{ID: mcIDs[0], QuestionNumber: 1, Type: domain.TypeMultipleChoice},
		{ID: mcIDs[1], QuestionNumber: 2, Type: domain.TypeMultipleChoice},
		{ID: esIDs[0], QuestionNumber: 3, Type: domain.TypeEssay},
		{ID: mcIDs[2], QuestionNumber: 4, Type: domain.TypeMultipleChoice},
		{ID: esIDs[1], QuestionNumber: 5, Type: domain.TypeEssay},
	}

	seed := GenerateSeed("student-abc", "schedule-xyz")
	result := ShuffleByGroup(questions, seed)

	if len(result) != 5 {
		t.Fatalf("Expected 5 questions, got %d", len(result))
	}

	// First 3 must all be MULTIPLE_CHOICE
	for i := 0; i < 3; i++ {
		if result[i].Type != domain.TypeMultipleChoice {
			t.Errorf("Position %d: expected MULTIPLE_CHOICE, got %s", i, result[i].Type)
		}
	}
	// Last 2 must all be ESSAY
	for i := 3; i < 5; i++ {
		if result[i].Type != domain.TypeEssay {
			t.Errorf("Position %d: expected ESSAY, got %s", i, result[i].Type)
		}
	}

	// Deterministic: same seed → same result
	result2 := ShuffleByGroup(questions, seed)
	for i := range result {
		if result[i].ID != result2[i].ID {
			t.Errorf("Not deterministic at index %d", i)
		}
	}
}
