package chat

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/lakshaysinghal/cligpt/internal/ai"
)

const systemPrompt = "You are CLIGpt, a concise and helpful assistant running inside a terminal."

// Session holds an in-memory conversation for the current run.
type Session struct {
	client   *ai.Client
	timeout  time.Duration
	messages []ai.Message
}

// New creates an empty session.
func New(client *ai.Client, timeout time.Duration) *Session {
	return &Session{
		client:  client,
		timeout: timeout,
	}
}

// Run drives the interactive REPL. It returns when the user exits or stdin closes.
func (s *Session) Run(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)
	printBanner()

	for {
		fmt.Print("\nyou > ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("\nbye!")
				return nil
			}
			return fmt.Errorf("read input: %w", err)
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		if isExit(input) {
			fmt.Println("bye!")
			return nil
		}
		if input == "/reset" {
			s.messages = nil
			fmt.Println("(conversation cleared)")
			continue
		}

		if err := s.ask(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
		}
	}
}

func (s *Session) ask(parent context.Context, input string) error {
	s.messages = append(s.messages, ai.Message{
		Role:    ai.RoleUser,
		Content: input,
	})

	ctx, cancel := context.WithTimeout(parent, s.timeout)
	defer cancel()

	fmt.Print("ai  > ")
	reply, err := s.client.Stream(ctx, systemPrompt, s.messages, func(chunk string) {
		fmt.Print(chunk)
	})
	fmt.Println()

	if err != nil {
		// Roll back the user message so a retry doesn't double-send it.
		s.messages = s.messages[:len(s.messages)-1]
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("request timed out after %s", s.timeout)
		}
		return err
	}

	s.messages = append(s.messages, ai.Message{
		Role:    ai.RoleModel,
		Content: reply,
	})
	return nil
}

func isExit(input string) bool {
	switch strings.ToLower(input) {
	case "/exit", "/quit", "exit", "quit":
		return true
	}
	return false
}

func printBanner() {
	fmt.Println("CLIGpt — type your message and press enter.")
	fmt.Println("Commands: /reset to clear history, /exit to quit.")
}