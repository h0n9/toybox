package main

import (
	"context"
	"fmt"
	"os/exec"

	ollama "github.com/ollama/ollama/api"
)

const (
	instruction = `Generate a concise and descriptive Git commit message based
	on the following changes: [Paste your diff here]. The changes are related to
	a new feature. Follow the 50/72 rule, use Conventional Commits format, and
	include a brief explanation of the changes in the body. Start with a clear,
	concise subject line.`
)

func main() {
	// init ctx
	ctx := context.Background()

	// run and read output: `git diff`
	cmd := exec.CommandContext(ctx, "git", "diff")
	gitDiffOutput, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// init ollamaClient
	ollamaClient, err := ollama.ClientFromEnvironment()
	if err != nil {
		panic(err)
	}

	// ask ai models (e.g., `ollama`) to write git commit messages with the diff
	stream := false
	ollamaReq := ollama.GenerateRequest{
		Model:  "llama3.1:8b",
		Prompt: string(gitDiffOutput),
		System: instruction,
		Stream: &stream,
	}
	err = ollamaClient.Generate(ctx, &ollamaReq, func(gr ollama.GenerateResponse) error {
		fmt.Println(gr.Response)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// print

	// TODO: write commit messages to `git commit`
}
