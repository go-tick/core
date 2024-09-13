package gotick

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarShouldCreateScheduleIfItsFuture(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Minute)

	s := NewCalendar(future)
	require.NotNil(t, s)

	assert.Equal(t, future.Format(time.RFC3339), s.Schedule())

	assert.Equal(t, future, s.First())

	assert.Equal(t, future, *s.Next(now))
	assert.Nil(t, s.Next(future))
	assert.Nil(t, s.Next(future.Add(1*time.Second)))
}
