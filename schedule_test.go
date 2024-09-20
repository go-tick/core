package gotick

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelayShouldWorkCorrectly(t *testing.T) {
	cron, err := NewCronSchedule("0 0 1 1 *")
	require.NoError(t, err)

	calendar := NewCalendarSchedule(time.Now().Add(1 * time.Minute))

	sequence, err := NewSequenceSchedule(time.Now(), time.Now().Add(1*time.Minute))
	require.NoError(t, err)

	data := []struct {
		name     string
		schedule JobSchedule
	}{
		{
			name:     "cron",
			schedule: cron,
		},
		{
			name:     "calendar",
			schedule: calendar,
		},
		{
			name:     "sequence",
			schedule: sequence,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			delay := NewJobScheduleWithMaxDelay(d.schedule, 1*time.Second)
			require.NotNil(t, delay)

			assert.Equal(t, d.schedule.Schedule(), delay.Schedule())
			assert.Equal(t, d.schedule.First(), delay.First())
			assert.Equal(t, d.schedule.Next(time.Now().Add(1*time.Minute)), delay.Next(time.Now().Add(1*time.Minute)))

			if md, ok := delay.(MaxDelay); ok {
				assert.Equal(t, 1*time.Second, md.MaxDelay())
			} else {
				assert.Fail(t, "delay should be a MaxDelay")
			}
		})
	}
}

func TestTimeoutShouldWorkCorrectly(t *testing.T) {
	cron, err := NewCronSchedule("0 0 1 1 *")
	require.NoError(t, err)

	calendar := NewCalendarSchedule(time.Now().Add(1 * time.Minute))

	sequence, err := NewSequenceSchedule(time.Now(), time.Now().Add(1*time.Minute))
	require.NoError(t, err)

	data := []struct {
		name     string
		schedule JobSchedule
	}{
		{
			name:     "cron",
			schedule: cron,
		},
		{
			name:     "calendar",
			schedule: calendar,
		},
		{
			name:     "sequence",
			schedule: sequence,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			timeout := NewJobScheduleWithTimeout(d.schedule, 1*time.Second)
			require.NotNil(t, timeout)

			assert.Equal(t, d.schedule.Schedule(), timeout.Schedule())
			assert.Equal(t, d.schedule.First(), timeout.First())
			assert.Equal(t, d.schedule.Next(time.Now().Add(1*time.Minute)), timeout.Next(time.Now().Add(1*time.Minute)))

			if mt, ok := timeout.(Timeout); ok {
				assert.Equal(t, 1*time.Second, mt.Timeout())
			} else {
				assert.Fail(t, "timeout should be a Timeout")
			}
		})
	}
}
