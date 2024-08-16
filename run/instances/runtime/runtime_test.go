package runtime

import (
	"context"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSyncMaker_MakeData(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	maker := CreateMaker(fndMock)
	data := maker.MakeData()

	assert.NotNil(t, data, "MakeData should return a non-nil Data instance")
	assert.IsType(t, &syncData{}, data, "The type of returned data should be *syncData")
}

func TestSyncMaker_MakeBackgroundContext(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	maker := CreateMaker(fndMock)
	ctx := maker.MakeBackgroundContext()

	assert.NotNil(t, ctx, "MakeBackgroundContext should return a non-nil context")
}

func TestSyncMaker_MakeContextWithTimeout_limited(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	maker := CreateMaker(fndMock)
	baseCtx := context.Background()
	timeout := 200 * time.Millisecond

	ctx, cancel := maker.MakeContextWithTimeout(baseCtx, timeout)
	defer cancel()

	assert.NotNil(t, ctx, "MakeContextWithTimeout should return a non-nil context")

	deadline, ok := ctx.Deadline()
	assert.True(t, ok, "Context should have a deadline")
	assert.WithinDuration(t, time.Now().Add(timeout), deadline, 50*time.Millisecond, "Deadline should be approximately now + timeout")

	select {
	case <-time.After(timeout + 100*time.Millisecond):
		assert.Fail(t, "Context should have been cancelled by timeout")
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err(), "Context should be cancelled due to deadline exceeded")
	}
}

func TestSyncMaker_MakeContextWithTimeout_unlimited(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	maker := CreateMaker(fndMock)
	baseCtx := context.Background()
	timeout := time.Millisecond * 0

	ctx, cancel := maker.MakeContextWithTimeout(baseCtx, timeout)
	defer cancel()

	select {
	case <-time.After(10 * time.Millisecond):
		assert.NoError(t, ctx.Err(), "Context should not have any error")
	case <-ctx.Done():
		assert.Fail(t, "Context should not have been cancelled")
	}
}
