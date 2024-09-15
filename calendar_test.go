package gotick

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarShouldCreateSchedule(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Minute)

	s := NewCalendarSchedule(future)
	require.NotNil(t, s)

	assert.Equal(t, future.Format(time.RFC3339), s.Schedule())

	assert.Equal(t, future, s.First())

	assert.Equal(t, future, *s.Next(now))
	assert.Nil(t, s.Next(future))
	assert.Nil(t, s.Next(future.Add(1*time.Second)))
}

func TestCalendarShouldCreateScheduleWithMaxDelay(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Minute)

	md := 1 * time.Second
	s := NewCalendarScheduleWithMaxDelay(future, md)
	require.NotNil(t, s)

	if smd, ok := s.(MaxDelay); ok {
		assert.Equal(t, md, smd.MaxDelay())
	} else {
		assert.Fail(t, "expected JobScheduleWithMaxDelay")
	}
}
