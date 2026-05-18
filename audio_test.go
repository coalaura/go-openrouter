package openrouter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSpeech_Basic(t *testing.T) {
	fakeClient := &fakeHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("audio-bytes")),
			Header: http.Header{
				"Content-Type":    []string{"audio/mpeg"},
				"X-Generation-Id": []string{"gen_123"},
			},
		},
	}

	cfg := DefaultConfig("test-token")
	cfg.BaseURL = "https://example.com/api/v1"
	cfg.HTTPClient = fakeClient

	client := NewClientWithConfig(*cfg)
	resp, err := client.CreateSpeech(context.Background(), SpeechRequest{
		Model:          "openai/gpt-4o-mini-tts",
		Input:          "Hello world",
		Voice:          "alloy",
		ResponseFormat: SpeechResponseFormatMp3,
		Speed:          1.25,
		Provider: &AudioProvider{
			Options: map[string]map[string]any{
				"openai": {
					"instructions": "Speak clearly.",
				},
			},
		},
	})
	require.NoError(t, err)

	require.NotNil(t, fakeClient.lastRequest)
	require.Equal(t, http.MethodPost, fakeClient.lastRequest.Method)
	require.Equal(t, "/api/v1/audio/speech", fakeClient.lastRequest.URL.Path)
	require.Equal(t, "audio/*", fakeClient.lastRequest.Header.Get("Accept"))
	require.Equal(t, "application/json; charset=utf-8", fakeClient.lastRequest.Header.Get("Content-Type"))
	require.Equal(t, "Bearer test-token", fakeClient.lastRequest.Header.Get("Authorization"))

	var body map[string]any
	require.NoError(t, json.NewDecoder(fakeClient.lastRequest.Body).Decode(&body))
	require.Equal(t, "openai/gpt-4o-mini-tts", body["model"])
	require.Equal(t, "Hello world", body["input"])
	require.Equal(t, "alloy", body["voice"])
	require.Equal(t, "mp3", body["response_format"])
	require.Equal(t, 1.25, body["speed"])
	require.Contains(t, body, "provider")

	require.Equal(t, []byte("audio-bytes"), resp.Audio)
	require.Equal(t, "audio/mpeg", resp.ContentType)
	require.Equal(t, "gen_123", resp.GenerationID)
}

func TestCreateSpeech_ErrorResponse(t *testing.T) {
	fakeClient := &fakeHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(strings.NewReader(`{"error":{"code":429,"message":"Rate limit exceeded"}}`)),
			Header:     make(http.Header),
		},
	}

	cfg := DefaultConfig("test-token")
	cfg.BaseURL = "https://example.com/api/v1"
	cfg.HTTPClient = fakeClient

	client := NewClientWithConfig(*cfg)
	_, err := client.CreateSpeech(context.Background(), SpeechRequest{
		Model: "openai/gpt-4o-mini-tts",
		Input: "Hello world",
		Voice: "alloy",
	})

	require.Error(t, err)
	require.True(t, IsErrorCode(err, http.StatusTooManyRequests))
}

func TestCreateTranscription_Basic(t *testing.T) {
	responseBody := `{
		"text": "Hello world",
		"usage": {
			"cost": 0.000508,
			"input_tokens": 83,
			"output_tokens": 30,
			"seconds": 9.2,
			"total_tokens": 113
		}
	}`
	fakeClient := &fakeHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Header:     make(http.Header),
		},
	}

	cfg := DefaultConfig("test-token")
	cfg.BaseURL = "https://example.com/api/v1"
	cfg.HTTPClient = fakeClient

	temperature := 0.0
	client := NewClientWithConfig(*cfg)
	resp, err := client.CreateTranscription(context.Background(), TranscriptionRequest{
		Model:       "openai/whisper-large-v3",
		InputAudio:  NewTranscriptionInputAudio([]byte("hello"), AudioFormatWav),
		Language:    "en",
		Temperature: &temperature,
	})
	require.NoError(t, err)

	require.NotNil(t, fakeClient.lastRequest)
	require.Equal(t, http.MethodPost, fakeClient.lastRequest.Method)
	require.Equal(t, "/api/v1/audio/transcriptions", fakeClient.lastRequest.URL.Path)
	require.Equal(t, "application/json; charset=utf-8", fakeClient.lastRequest.Header.Get("Accept"))
	require.Equal(t, "application/json; charset=utf-8", fakeClient.lastRequest.Header.Get("Content-Type"))

	var body struct {
		Model       string                  `json:"model"`
		InputAudio  TranscriptionInputAudio `json:"input_audio"`
		Language    string                  `json:"language"`
		Temperature *float64                `json:"temperature"`
	}
	require.NoError(t, json.NewDecoder(fakeClient.lastRequest.Body).Decode(&body))
	require.Equal(t, "openai/whisper-large-v3", body.Model)
	require.Equal(t, "aGVsbG8=", body.InputAudio.Data)
	require.Equal(t, AudioFormatWav, body.InputAudio.Format)
	require.Equal(t, "en", body.Language)
	require.NotNil(t, body.Temperature)
	require.Equal(t, 0.0, *body.Temperature)

	require.Equal(t, "Hello world", resp.Text)
	require.NotNil(t, resp.Usage)
	require.InDelta(t, 0.000508, resp.Usage.Cost, 1e-12)
	require.Equal(t, 83, resp.Usage.InputTokens)
	require.Equal(t, 30, resp.Usage.OutputTokens)
	require.InDelta(t, 9.2, resp.Usage.Seconds, 1e-12)
	require.Equal(t, 113, resp.Usage.TotalTokens)
}

func TestNewTranscriptionInputAudioFromFileUnsupportedFormat(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "voice.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello"), 0o600))

	_, err := NewTranscriptionInputAudioFromFile(filePath)
	require.EqualError(t, err, "unsupported audio format: .txt")
}
