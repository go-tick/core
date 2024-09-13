package gotick

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceShouldCreateSchedule(t *testing.T) {
	time1 := time.Now()
	time2 := time1.Add(1 * time.Minute)
	s, err := NewSequenceSchedule(time1, time2)

	require.NoError(t, err)
	require.NotNil(t, s)

	expectedSchedule := fmt.Sprintf("%s,%s", time1.Format(time.RFC3339), time2.Format(time.RFC3339))
	assert.Equal(t, expectedSchedule, s.Schedule())

	assert.Equal(t, time1, s.First())
	assert.Equal(t, time1, s.First()) // First should not change the state

	assert.Equal(t, time1, *s.Next(time1.Add(-1 * time.Second)))
	assert.Equal(t, time2, *s.Next(time1))
	assert.Nil(t, s.Next(time2))
	assert.Nil(t, s.Next(time2.Add(1*time.Second)))
}

func TestSequenceShouldFailToCreateScheduleWithInvalidSequence(t *testing.T) {
	time1 := time.Now()
	time2 := time1.Add(1 * time.Minute)
	_, err := NewSequenceSchedule(time2, time1)

	require.Error(t, err)
	assert.Equal(t, ErrInvalidSequenceSchedule, err)
}

func TestSequenceShouldCreateScheduleWithMaxDelay(t *testing.T) {
	time1 := time.Now()
	time2 := time1.Add(1 * time.Minute)
	md := 1 * time.Second
	s, err := NewSequenceWithMaxDelay(md, time1, time2)

	require.NoError(t, err)
	require.NotNil(t, s)

	if smd, ok := s.(JobScheduleWithMaxDelay); ok {
		assert.Equal(t, md, *smd.MaxDelay())
	} else {
		assert.Fail(t, "expected JobScheduleWithMaxDelay")
	}
}
