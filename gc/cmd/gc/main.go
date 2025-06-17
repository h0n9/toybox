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

	// retrieve last 3 lates commit messages
	cmd := exec.CommandContext(ctx, "git", "log", "-3", "--pretty=format:'''\n%B\n'''\n")
	gitLogOutput, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// run and read output: `git diff --staged`
	cmd = exec.CommandContext(ctx, "git", "diff", "--staged", "--diff-algorithm=minimal")
	gitDiffOutput, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// init ollamaClient
	ollamaClient, err := ollama.ClientFromEnvironment()
	if err != nil {
		panic(err)
	}

	stream := false
	// generate dynamic prompt
	prompt := fmt.Sprintf(instruction, string(gitLogOutput), string(gitDiffOutput))

	// ask ai models (e.g., `ollama`) to write git commit messages
	ollamaReq := ollama.GenerateRequest{
		Model:  "llama3.1:8b",
		Prompt: prompt,
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
