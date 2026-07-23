// Package operation defines format-neutral file safety and publication
// contracts. Concrete implementations are future V1 capabilities.
package operation

import (
	"context"
	"io"
	"time"
)

type SourceRef struct {
	ID       string
	Location string
}

type Snapshot struct {
	SourceID string
	Location string
	Digest   string
}

type Snapshotter interface {
	Snapshot(context.Context, SourceRef) (Snapshot, error)
}

type LockRequest struct {
	Location   string
	Retries    int
	Interval   time.Duration
	Timeout    time.Duration
	SharedRead bool
}

type Lock interface {
	Release() error
}

type Locker interface {
	Acquire(context.Context, LockRequest) (Lock, error)
}

type StageRequest struct {
	TargetLocation string
}

type Stage struct {
	Location string
	Writer   io.WriteCloser
}

type Stager interface {
	Stage(context.Context, StageRequest) (Stage, error)
}

type Artifact struct {
	Location string
	Digest   string
}

type Validation struct {
	Passed  bool
	Message string
}

type Validator interface {
	Validate(context.Context, Artifact) (Validation, error)
}

type PublishRequest struct {
	Staged         Artifact
	TargetLocation string
	Collision      string
}

type Publisher interface {
	Publish(context.Context, PublishRequest) (Artifact, error)
}

type Workspace struct {
	Location string
	Success  bool
}

type Cleaner interface {
	Cleanup(context.Context, Workspace) error
}
