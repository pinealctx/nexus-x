package agentic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"charm.land/fantasy"
	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"
)

func TestRunLLMStreamErrorTerminatesCurrentStreamOnly(t *testing.T) {
	streamErr := errors.New("provider unavailable")
	agent := failingStreamAgent{err: streamErr}
	channel := &recordingStreamingChannel{}
	errorHandlerCalled := false

	engine, err := NewEngine(
		WithAgent(agent),
		WithRouter(NewRouter()),
		WithStreamMode(),
		WithErrorHandler(func(context.Context, *IncomingUpdate, error) error {
			errorHandlerCalled = true
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	ctx := WithChannel(context.Background(), channel)
	update := &IncomingUpdate{
		Channel:        channel,
		UserID:         42,
		ConversationID: 1001,
		Text:           "hello",
	}

	if err := engine.RunLLM(ctx, update); err != nil {
		t.Fatalf("RunLLM() error = %v", err)
	}

	if channel.startCalls != 1 {
		t.Fatalf("StartStream calls = %d, want 1", channel.startCalls)
	}
	if channel.writer.errorCalls != 1 {
		t.Fatalf("StreamWriter.Error calls = %d, want 1", channel.writer.errorCalls)
	}
	if !strings.Contains(channel.writer.errorMessage, streamErr.Error()) {
		t.Fatalf("stream error message %q does not contain %q", channel.writer.errorMessage, streamErr.Error())
	}
	if errorHandlerCalled {
		t.Fatal("error handler was called after stream error was already written")
	}
	if channel.sendCalls != 0 {
		t.Fatalf("SendMessage calls = %d, want 0", channel.sendCalls)
	}
}

type failingStreamAgent struct {
	err error
}

func (a failingStreamAgent) Generate(context.Context, fantasy.AgentCall) (*fantasy.AgentResult, error) {
	return nil, a.err
}

func (a failingStreamAgent) Stream(context.Context, fantasy.AgentStreamCall) (*fantasy.AgentResult, error) {
	return nil, a.err
}

type recordingStreamingChannel struct {
	startCalls int
	sendCalls  int
	writer     recordingStreamWriter
}

func (c *recordingStreamingChannel) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResult, error) {
	c.sendCalls++
	return &SendMessageResult{MessageID: int64(c.sendCalls)}, nil
}

func (c *recordingStreamingChannel) EditMessage(context.Context, int64, int64, *sharedv1.MessageBody) error {
	return nil
}

func (c *recordingStreamingChannel) RecallMessage(context.Context, int64, int64) error {
	return nil
}

func (c *recordingStreamingChannel) AnswerCardAction(context.Context, int64, int64, string, string, bool) error {
	return nil
}

func (c *recordingStreamingChannel) StartStream(context.Context, int64) (StreamWriter, error) {
	c.startCalls++
	return &c.writer, nil
}

type recordingStreamWriter struct {
	pushCalls    int
	endCalls     int
	errorCalls   int
	errorMessage string
}

func (w *recordingStreamWriter) Push(context.Context, string) error {
	w.pushCalls++
	return nil
}

func (w *recordingStreamWriter) End(context.Context, string) error {
	w.endCalls++
	return nil
}

func (w *recordingStreamWriter) Error(_ context.Context, msg string) error {
	w.errorCalls++
	w.errorMessage = msg
	return nil
}
