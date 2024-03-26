package revai

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMetadata = "test-metadata"
const testMediaURL = "https://support.rev.com/hc/en-us/article_attachments/200043975/FTC_Sample_1_-_Single.mp3"

func TestJobService_SubmitFile(t *testing.T) {
	f, err := os.Open("./testdata/img.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	params := &NewFileJobParams{
		Media:    f,
		Filename: f.Name(),
	}

	ctx := context.Background()

	newJob, err := testClient.Job.SubmitFile(ctx, params)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, newJob.ID, "new job id should not be nil")
	assert.Equal(t, "in_progress", newJob.Status, "response status should be in_progress")
}

type transcriptionResponse struct {
	JobID   string
	Caption TranscriptionCaption
	Text    TranscriptionText
}

type TranscriptionText string
type TranscriptionCaption struct {
	// .srt file as text
	SubripText string `bson:"subrip_text"`
	WebVTT     string `bson:"web_vtt"`
}

func TestJobService_SubmitFileWithOptionTranslate(t *testing.T) {
	f, err := os.Open("./testdata/img.jpg")
	require.NoError(t, err)

	defer f.Close()

	params := &NewFileJobParams{
		Media:    f,
		Filename: f.Name(),
		JobOptions: &JobOptions{
			Metadata: testMetadata,
			Language: "en", // revai support said at the moment to translate origin language is required
			TranslationConfig: JobOptionTranslationConfig{
				TargetLanguages: []JobOptionTranslationConfigTargetLanguages{
					{
						Language: "es",
						Model:    "premium",
					},
				},
			},
		},
	}

	ctx := context.Background()

	newJob, err := testClient.Job.SubmitFile(ctx, params)
	require.NoError(t, err)

	assert.NotNil(t, newJob.ID, "new job id should not be nil")
	assert.Equal(t, testMetadata, newJob.Metadata, "meta data should be set")
	assert.Equal(t, "in_progress", newJob.Status, "response status should be in_progress")

	job, err := testClient.Job.Get(ctx, &GetJobParams{
		ID: testJob.ID,
	})
	require.NoError(t, err)

	errChan := make(chan error, 1)

	go func(job *Job, errChan chan<- error) {
		//var transcript *revai.TextTranscript
		for {
			job, err = testClient.Job.Get(ctx, &GetJobParams{
				ID: job.ID,
			})
			if err != nil {
				errChan <- fmt.Errorf("revaiClient.Transcript.Get: %s", err)
				return
			}

			if job.Status == "failed" {
				errChan <- fmt.Errorf("job.Status: %s | job.Failure: %s | job.FailureDetail: %s", job.Status, job.Failure, job.FailureDetail)
				return
			}

			if job.Status == "transcribed" {
				var wg sync.WaitGroup
				var tr transcriptionResponse
				tr.JobID = job.ID

				wg.Add(3)
				go func(tr *transcriptionResponse) {
					defer wg.Done()
					//job.MediaURL
					plainText, err := testClient.Transcript.GetText(ctx, &GetTranscriptParams{
						JobID: job.ID,
					})
					if err != nil {
						errChan <- err
						return
					}
					tr.Text = TranscriptionText(plainText.Value)
				}(&tr)

				go func(tr *transcriptionResponse) {
					defer wg.Done()
					srtCaption, err := testClient.Caption.Get(ctx, &GetCaptionParams{
						JobID:  job.ID,
						Accept: "application/x-subrip",
					})
					if err != nil {
						errChan <- err
						return
					}
					tr.Caption.SubripText = srtCaption.Value

				}(&tr)

				go func(tr *transcriptionResponse) {
					defer wg.Done()
					vttCaption, err := testClient.Caption.Get(ctx, &GetCaptionParams{
						JobID:  job.ID,
						Accept: "text/vtt",
					})
					if err != nil {
						errChan <- err
						return
					}

					tr.Caption.WebVTT = vttCaption.Value
				}(&tr)

				wg.Wait()
				defer func() { errChan <- nil }()
				t.Log("tr.Caption.SubripText", tr.Caption.SubripText)
				t.Log("tr.Caption.SubripText", tr.Caption.WebVTT)

				require.NotEmpty(t, tr.Caption.SubripText)
				require.NotEmpty(t, tr.Caption.WebVTT)

				return
			}
			time.Sleep(time.Millisecond * 300)
		}
	}(job, errChan)

	require.NoError(t, <-errChan)

	assert.NotNil(t, newJob.ID, "new job id should not be nil")

}

func TestJobService_SubmitFileWithOption(t *testing.T) {
	f, err := os.Open("./testdata/testaudio.mp3")
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	params := &NewFileJobParams{
		Media:    f,
		Filename: f.Name(),
		JobOptions: &JobOptions{
			Metadata: testMetadata,
			CustomVocabularies: []JobOptionCustomVocabulary{
				{
					Phrases: []string{
						"Paul McCartney",
						"Amelia Earhart",
						"Weiss-Bergman",
						"BLM",
					},
				},
			},
		},
	}

	ctx := context.Background()

	newJob, err := testClient.Job.SubmitFile(ctx, params)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, newJob.ID, "new job id should not be nil")
	assert.Equal(t, testMetadata, newJob.Metadata, "meta data should be set")
	assert.Equal(t, "in_progress", newJob.Status, "response status should be in_progress")
}

func TestJobService_SubmitURL(t *testing.T) {
	params := &NewURLJobParams{
		MediaURL: testMediaURL,
	}

	ctx := context.Background()

	newJob, err := testClient.Job.SubmitURL(ctx, params)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, newJob.ID, "new job id should not be nil")
	assert.Equal(t, "in_progress", newJob.Status, "response status should be in_progress")
}

func TestJobService_SubmitWithOption(t *testing.T) {
	params := &NewURLJobParams{
		MediaURL: testMediaURL,
		Metadata: testMetadata,
	}

	ctx := context.Background()

	newJob, err := testClient.Job.SubmitURL(ctx, params)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, newJob.ID, "new job id should not be nil")
	assert.Equal(t, testMetadata, newJob.Metadata, "meta data should be set")
	assert.Equal(t, "in_progress", newJob.Status, "response status should be in_progress")
}

func TestJobService_Get(t *testing.T) {
	params := &GetJobParams{
		ID: testJob.ID,
	}

	ctx := context.Background()

	newJob, err := testClient.Job.Get(ctx, params)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, newJob.ID, "new job id should not be nil")
}

func TestJobService_Delete(t *testing.T) {
	deletableJob := makeTestJob()

	params := &DeleteJobParams{
		ID: deletableJob.ID,
	}

	ctx := context.Background()

	if err := testClient.Job.Delete(ctx, params); err != nil {
		t.Error(err)
		return
	}
}

func TestJobService_List(t *testing.T) {
	params := &ListJobParams{}

	ctx := context.Background()

	jobs, err := testClient.Job.List(ctx, params)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, jobs, "jobs should not be nil")
}

func TestJobService_ListWithLimit(t *testing.T) {
	params := &ListJobParams{
		Limit: 2,
	}

	ctx := context.Background()

	jobs, err := testClient.Job.List(ctx, params)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, 2, len(jobs), "it returns 2 jobs when limit is set to 2")
}
