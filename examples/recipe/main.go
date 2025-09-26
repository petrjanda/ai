package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsynq/cloud/ai-data-sre/examples"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/workflows"
	"github.com/joho/godotenv"
)

// Define input schema
type RecipeInput struct {
	Ingredient          string `json:"ingredient" jsonschema:"description=Main ingredient or cuisine type"`
	DietaryRestrictions string `json:"dietaryRestrictions,omitempty" jsonschema:"description=Any dietary restrictions"`
}

// Define output schema
type Recipe struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	PrepTime     string   `json:"prepTime"`
	CookTime     string   `json:"cookTime"`
	Servings     int      `json:"servings"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
	Tips         []string `json:"tips,omitempty"`
}

func main() {
	ctx := context.Background()
	godotenv.Load()

	llm := examples.GetLiteLLM()

	flow := workflows.NewTypedTask[Recipe]("recipe", ai.NewLLMRequest(
		ai.WithModel(ai.Gemini25Flash),
		ai.WithSystem(`
				Generate a recipe based on the user's input which includes a given ingredient and dietary restrictions.
			`),
	))

	response, err := flow.Invoke(ctx, llm, ai.NewHistory(
		ai.NewUserMessage("Ingridient: avocado, pastas, chicken, Dietary Restrictions: none"),
	))

	if err != nil {
		panic(err)
	}

	payload, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(payload))
}
