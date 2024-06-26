package revai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// TranscriptService provides access to the transcript related functions
// in the Rev.ai API.
type TranscriptService service

// Transcript represents a Rev.ai job json transcript
type Transcript struct {
	Monologues []Monologue `json:"monologues"`
}

// Monologue represents a Rev.ai monologue
type Monologue struct {
	Speaker  int       `json:"speaker"`
	Elements []Element `json:"elements"`
}

// Element represents a Rev.ai element
type Element struct {
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Ts         float64 `json:"ts"`
	EndTs      float64 `json:"end_ts"`
	Confidence float64 `json:"confidence"`
}

// GetTranscriptParams specifies the parameters to the
// TranscriptService.Get method.
type GetTranscriptParams struct {
	JobID string
}

// Get returns the transcript for a completed transcription job in JSON format.
// https://www.rev.ai/docs#operation/GetTranscriptById
func (s *TranscriptService) Get(ctx context.Context, params *GetTranscriptParams) (*Transcript, error) {
	urlPath := "/speechtotext/v1/jobs/" + params.JobID + "/transcript"

	req, err := s.client.newRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request %w", err)
	}

	req.Header.Add("Accept", revTranscriptJSONHeader)

	var transcript Transcript
	if err := s.client.doJSON(ctx, req, &transcript); err != nil {
		return nil, err
	}

	return &transcript, nil
}

// TextTranscript represents a Rev.ai job text transcript
type TextTranscript struct {
	Value string
}

// Get returns the transcript for a completed transcription job in text format.
// https://www.rev.ai/docs#operation/GetTranscriptById
func (s *TranscriptService) GetText(ctx context.Context, params *GetTranscriptParams) (*TextTranscript, error) {
	urlPath := "/speechtotext/v1/jobs/" + params.JobID + "/transcript"

	req, err := s.client.newRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request %w", err)
	}

	req.Header.Add("Accept", textPlainHeader)

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}

	transcript := TextTranscript{
		Value: buf.String(),
	}

	return &transcript, nil
}

// GetTranslation returns the translated transcript on language sepcified for a completed transcription job in text format.
// https://docs.rev.ai/api/asynchronous/reference/#operation/GetTranscriptById
func (s *TranscriptService) GetTranslation(ctx context.Context, targetLanguage string, params *GetTranscriptParams) (*TextTranscript, error) {
	urlPath := "/speechtotext/v1/jobs/" + params.JobID + "/transcript/translation" + targetLanguage

	req, err := s.client.newRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request %w", err)
	}

	req.Header.Add("Accept", textPlainHeader)

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}

	transcript := TextTranscript{
		Value: buf.String(),
	}

	return &transcript, nil
}
