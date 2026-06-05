package models

import "time"

// AdminOrphanFileScan represents the latest orphan-file scan snapshot.
type AdminOrphanFileScan struct {
	ID              int64                    `json:"id"`
	Status          string                   `json:"status"`
	TotalFiles      int64                    `json:"total_files"`
	ReferencedFiles int64                    `json:"referenced_files"`
	OrphanFiles     int64                    `json:"orphan_files"`
	DeletedFiles    int64                    `json:"deleted_files"`
	Error           string                   `json:"error"`
	StartedAt       *time.Time               `json:"started_at"`
	FinishedAt      *time.Time               `json:"finished_at"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
	Items           []AdminOrphanFileScanItem `json:"items"`
}

// AdminOrphanFileScanItem is a single orphan file candidate.
type AdminOrphanFileScanItem struct {
	ID            int64     `json:"id"`
	FilePath      string    `json:"file_path"`
	RelativePath  string    `json:"relative_path"`
	SizeBytes     int64     `json:"size_bytes"`
	ModTime       time.Time `json:"mod_time"`
	CreatedAt     time.Time `json:"created_at"`
}
