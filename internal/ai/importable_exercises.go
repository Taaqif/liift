package ai

import (
	_ "embed"
	"encoding/json"
	"log"
	"sync"
)

//go:embed exercises.json
var importableExercisesJSON []byte

type importableExercise struct {
	Name           string   `json:"name"`
	Force          *string  `json:"force"`
	Equipment      *string  `json:"equipment"`
	PrimaryMuscles []string `json:"primaryMuscles"`
	Category       string   `json:"category"`
}

var (
	importableOnce      sync.Once
	importableExercises []importableExercise
)

func getImportableExercises() []importableExercise {
	importableOnce.Do(func() {
		if err := json.Unmarshal(importableExercisesJSON, &importableExercises); err != nil {
			log.Printf("Failed to parse importable exercises: %v", err)
		}
	})
	return importableExercises
}
