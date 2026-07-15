package llmservice

import (
	"errors"
	"fmt"

	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
)

func HandleIDEFallback(err error, command string, tasks []pipeline.PendingTask, workDir string) (string, error) {
	if errors.Is(err, ollama.ErrIDEFallback) {
		taskListPath, writeErr := pipeline.WriteTaskList(command, tasks)
		if writeErr != nil {
			return "", fmt.Errorf("failed to write task list: %v", writeErr)
		}

		msg := fmt.Sprintf("Task delegated to IDE Agent: %s\n", taskListPath)
		if workDir != "" {
			msg += fmt.Sprintf("Working directory: %s\n", workDir)
		}
		msg += "Run `atd continue` to resume."
		fmt.Println(msg)
		return msg, nil
	}
	return "", err
}

func HandleSimpleIDEFallback(err error, command, prompt, outputSchema string) (string, error) {
	if errors.Is(err, ollama.ErrIDEFallback) {
		promptName := "task_" + command
		resultFile := promptName + ".result"

		promptPath, writeErr := pipeline.WritePromptFile(promptName, prompt)
		if writeErr != nil {
			return "", fmt.Errorf("failed to write prompt file: %v", writeErr)
		}

		tasks := []pipeline.PendingTask{
			{
				PromptFile:   promptPath,
				ResultFile:   resultFile,
				Instruction:  command,
				OutputSchema: outputSchema,
			},
		}

		return HandleIDEFallback(err, command, tasks, "")
	}
	return "", err
}

type SimpleTask struct {
	Command      string
	Prompt       string
	OutputSchema string
}

func HandleSimpleIDEFallbackMulti(err error, command string, tasks []SimpleTask) (string, error) {
	if errors.Is(err, ollama.ErrIDEFallback) {
		var pipelineTasks []pipeline.PendingTask
		for i, task := range tasks {
			promptName := fmt.Sprintf("%s_step_%d", command, i+1)
			resultFile := fmt.Sprintf("%s_step_%d.result", command, i+1)

			promptPath, writeErr := pipeline.WritePromptFile(promptName, task.Prompt)
			if writeErr != nil {
				return "", fmt.Errorf("failed to write prompt file for step %d: %v", i+1, writeErr)
			}

			pipelineTasks = append(pipelineTasks, pipeline.PendingTask{
				PromptFile:   promptPath,
				ResultFile:   resultFile,
				Instruction:  task.Command,
				OutputSchema: task.OutputSchema,
			})
		}

		return HandleIDEFallback(err, command, pipelineTasks, "")
	}
	return "", err
}