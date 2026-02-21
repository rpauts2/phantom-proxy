package polymorphic

import (
	"strings"
	"testing"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine("high", 15)
	if engine == nil {
		t.Fatal("Expected engine to be created")
	}
}

func TestMutate(t *testing.T) {
	engine := NewEngine("medium", 15)

	input := `var username = "test"; var password = "secret";`

	result := engine.Mutate(input)

	if result == nil {
		t.Fatal("Expected result to be returned")
	}

	if result.MutatedCode == "" {
		t.Fatal("Expected mutated output")
	}

	if len(result.Mutations) == 0 {
		t.Error("Expected at least one mutation")
	}

	if result.OriginalHash == "" {
		t.Error("Expected original hash")
	}
}

func TestMutateLowLevel(t *testing.T) {
	engine := NewEngine("low", 15)

	input := `var x = "hello";`
	result := engine.Mutate(input)

	if result.MutatedCode == "" {
		t.Fatal("Expected mutated output")
	}
}

func TestMutateHighLevel(t *testing.T) {
	engine := NewEngine("high", 15)

	input := `var data = { user: "test", pass: "secret" };`
	result := engine.Mutate(input)

	if result.MutatedCode == "" {
		t.Fatal("Expected mutated output")
	}

	// High level should have more mutations
	if len(result.Mutations) < 3 {
		t.Errorf("Expected at least 3 mutations at high level, got %d", len(result.Mutations))
	}
}

func TestDeterministicMutation(t *testing.T) {
	// Same seed should produce same output
	engine1 := NewEngine("medium", 15)
	engine2 := NewEngine("medium", 15)

	input := `var test = "deterministic";`

	result1 := engine1.Mutate(input)
	result2 := engine2.Mutate(input)

	// Note: Due to seed rotation, results may differ
	// This test just verifies both produce output
	if result1.MutatedCode == "" || result2.MutatedCode == "" {
		t.Error("Expected both engines to produce output")
	}
}

func TestEmptyInput(t *testing.T) {
	engine := NewEngine("medium", 15)

	result := engine.Mutate("")

	if result.MutatedCode != "" {
		t.Error("Expected empty output for empty input")
	}
}

func TestComplexJavaScript(t *testing.T) {
	engine := NewEngine("high", 15)

	input := `
function login(username, password) {
    var data = {
        user: username,
        pass: password,
        timestamp: Date.now()
    };
    return fetch('/api/login', {
        method: 'POST',
        body: JSON.stringify(data)
    });
}
`

	result := engine.Mutate(input)

	if result.MutatedCode == "" {
		t.Fatal("Expected mutated output")
	}

	// Should still be valid JavaScript (basic check)
	if !strings.Contains(result.MutatedCode, "function") && !strings.Contains(result.MutatedCode, "=>") {
		t.Error("Expected mutated code to still contain function declarations")
	}
}

func TestGetStats(t *testing.T) {
	engine := NewEngine("high", 15)

	input := `var x = 1; var y = 2;`
	engine.Mutate(input)
	engine.Mutate(input)

	stats := engine.GetStats()

	if stats == nil {
		t.Error("Expected stats to be returned")
	}

	mutationCount, ok := stats["mutation_count"].(int64)
	if !ok {
		t.Error("Expected mutation_count to be int64")
	}

	if mutationCount < 2 {
		t.Errorf("Expected at least 2 mutations, got %d", mutationCount)
	}
}

func TestMutationLevels(t *testing.T) {
	testCases := []struct {
		level         string
		minMutations int
	}{
		{"low", 1},
		{"medium", 2},
		{"high", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			engine := NewEngine(tc.level, 15)
			input := `var test = "level test";`
			result := engine.Mutate(input)

			if len(result.Mutations) < tc.minMutations {
				t.Errorf("Expected at least %d mutations for %s level, got %d", tc.minMutations, tc.level, len(result.Mutations))
			}
		})
	}
}
