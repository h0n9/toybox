package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	ollama "github.com/ollama/ollama/api"
)

const (
	instruction = `
Generate a git commit message following this structure:
1. First line: conventional commit format under 50 characters (type: concise
description) (remember to use semantic types like feat, fix, docs, style,
refactor, perf, test, chore, etc.)
2. Optional bullet points under 72 characters for each line if more context
helps:
   - Keep the second line blank
   - Keep them short and direct
   - Focus on what changed
   - Always be terse
   - Don't overly explain
   - Drop any fluffy or formal language

Keep in mind that you should not repeat the same message but should write the
ones based on the +/- of codes. Also, the first line message should be brief but
should contain what changed.

Return ONLY the commit message - no introduction, no explanation, no quotes
around it.

Examples:
'''
feat: Add user auth system

- Add JWT tokens for API auth
- Handle token refresh for long sessions
'''

'''
fix: Resolve memory leak in worker pool

- Clean up idle connections
- Add timeout for stale workers
'''

Simple change example:
'''
fix: Typo in README.md
'''

Very important: Do not respond with any of the examples. Your message must be
based off the diff that is about to be provided, with a little bit of styling
informed by the recent commits you're about to see.

Recent commits from this repo (for style reference):
%s

Here's the diff:

%s
`

	sampleGitLogOutput = `feat: Add user authentication with JWT

This commit introduces a new user authentication system using JSON Web Tokens.
It includes the necessary API endpoints, token generation logic, and validation middleware.
		`
)

func main() {
	// init ctx
	ctx := context.Background()

	// retrieve last 3 lates commit messages
	cmd := exec.CommandContext(ctx, "git", "log", "-3", "--pretty=format:'''\n%B\n'''\n")
	gitLogOutput, err := cmd.Output()
	if err != nil {
		// fallback on error for retieving latest commits
		gitLogOutput = []byte(sampleGitLogOutput)
	}

	// run and read output: `git diff --staged`
	cmd = exec.CommandContext(ctx, "git", "diff", "--staged", "--diff-algorithm=minimal")
	gitDiffOutput, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	if len(gitDiffOutput) == 0 {
		fmt.Println("gc: found nothing to commit")
		os.Exit(0)
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
