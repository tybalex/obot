package types

type CronJob struct {
	Metadata
	CronJobManifest
	LastRunStartedAt           *Time `json:"lastRunStartedAt,omitempty"`
	LastSuccessfulRunCompleted *Time `json:"lastSuccessfulRunCompleted,omitempty"`
	NextRunAt                  *Time `json:"nextRunAt,omitempty"`
}

type CronJobManifest struct {
	Description  string    `json:"description,omitempty"`
	Schedule     string    `json:"schedule,omitempty"`
	Workflow     string    `json:"workflow,omitempty"`
	Input        string    `json:"input,omitempty"`
	UserID       string    `json:"userID,omitempty"`
	TaskSchedule *Schedule `json:"taskSchedule,omitempty"`
}

type CronJobList List[CronJob]
