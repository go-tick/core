package gotick

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarShouldNotCreateScheduleIfItsPast(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Millisecond)

	_, err := NewCalendar(past)

	require.Error(t, err)
	assert.Equal(t, ErrPastTime, err)
}

func TestCalendarShouldCreateScheduleIfItsFuture(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Minute)

	s, err := NewCalendar(future)

	require.NoError(t, err)
	require.NotNil(t, s)

	assert.Equal(t, future.Format(time.RFC3339), s.Schedule())

	assert.Equal(t, future, *s.Next(now))
	assert.Equal(t, future, *s.Next(future))
	assert.Nil(t, s.Next(future.Add(1*time.Second)))
}
