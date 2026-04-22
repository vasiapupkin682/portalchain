package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type TaskStatus int32

const (
    TaskStatusPending   TaskStatus = 0
    TaskStatusAssigned  TaskStatus = 1
    TaskStatusCompleted TaskStatus = 2
    TaskStatusExpired   TaskStatus = 3
    TaskStatusDisputed  TaskStatus = 4 // reserved for future
)

type Task struct {
    ID              string     `json:"id"`
    Creator         string     `json:"creator"`
    QueryHash       string     `json:"query_hash"`
    QueryURL        string     `json:"query_url"`
    TaskType        string     `json:"task_type"`
    Reward          sdk.Coin   `json:"reward"`
    FreeQuota       bool       `json:"free_quota"`
    Agent           string     `json:"agent"`
    Status          TaskStatus `json:"status"`
    CreatedAt       int64      `json:"created_at"`
    Deadline        int64      `json:"deadline"`
    ResultHash      string     `json:"result_hash"`
    ResultURL       string     `json:"result_url"`
    DisputeDeadline int64      `json:"dispute_deadline"` // reserved for future
}

type DailyQuota struct {
    Address       string `json:"address"`
    Date          string `json:"date"` // YYYY-MM-DD
    TextCount     int64  `json:"text_count"`
    CodeCount     int64  `json:"code_count"`
    AnalysisCount int64  `json:"analysis_count"`
}
