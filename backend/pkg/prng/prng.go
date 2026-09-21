package prng

import (
	"hash/fnv"
	"math/rand"

	"cbt-backend/internal/domain"
)

// GenerateSeed computes a deterministic 64-bit integer seed from studentID and scheduleID
func GenerateSeed(studentID, scheduleID string) int64 {
	h := fnv.New64a()
	h.Write([]byte(studentID))
	h.Write([]byte(":"))
	h.Write([]byte(scheduleID))
	return int64(h.Sum64())
}

// GenerateSubSeed creates a deterministic seed for a specific question given the master seed
func GenerateSubSeed(masterSeed int64, questionID string) int64 {
	h := fnv.New64a()
	h.Write([]byte(questionID))
	return masterSeed ^ int64(h.Sum64())
}

// ShuffleQuestions returns a shuffled copy of the questions slice using the seed
func ShuffleQuestions(questions []domain.Question, seed int64) []domain.Question {
	if len(questions) <= 1 {
		return questions
	}
	r := rand.New(rand.NewSource(seed))
	shuffled := make([]domain.Question, len(questions))
	copy(shuffled, questions)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

// ShuffleOptions returns a shuffled copy of OptionItems using a sub-seed
func ShuffleOptions(options []domain.OptionItem, subSeed int64) []domain.OptionItem {
	if len(options) <= 1 {
		return options
	}
	r := rand.New(rand.NewSource(subSeed))
	shuffled := make([]domain.OptionItem, len(options))
	copy(shuffled, options)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}
