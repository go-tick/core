package gotick

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validCrons = []struct {
	name     string
	schedule string
}{
	{
		name:     "every minute",
		schedule: "* * * * *",
	},
	{
		name:     "at minute 0",
		schedule: "0 * * * *",
	},
	{
		name:     "at minute 0 and hour 0",
		schedule: "0 0 * * *",
	},
	{
		name:     "at minute 0 and hour 0 and day 1",
		schedule: "0 0 1 * *",
	},
	{
		name:     "at minute 0 and hour 0 and day 1 and month 1",
		schedule: "0 0 1 1 *",
	},
	{
		name:     "at minute 0 and hour 0 and day 1 and month 1 and day of week 0",
		schedule: "0 0 1 1 0",
	},
	{
		name:     "at minute 0 and hour 0 and day 1 and month 1 and Monday",
		schedule: "0 0 1 1 MON",
	},
	{
		name:     "at minute 0 past hour 0 and 12 on day-of-month 1 in every 2nd month",
		schedule: "0 0,12 1 */2 *",
	},
}

func TestShouldNotCreateCronIfItsInvalid(t *testing.T) {
	data := []struct {
		name     string
		schedule string
	}{
		{
			name:     "empty",
			schedule: "",
		},
		{
			name:     "invalid",
			schedule: "invalid",
		},
		{
			name:     "invalid cron",
			schedule: "0 0 0 0 0",
		},
		{
			name:     "invalid cron 2",
			schedule: "0 0 0 0 0 0",
		},
		{
			name:     "at minute 0 and hour 0 and day 0",
			schedule: "0 0 0 * *",
		},
		{
			name:     "at minute 0 and hour 0 and day 0 and month 0",
			schedule: "0 0 1 0 *",
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			_, err := NewCronSchedule(d.schedule)

			require.Error(t, err)
			assert.Equal(t, err, ErrInvalidCron)
		})
	}
}

func TestShouldCreateCron(t *testing.T) {
	for _, d := range validCrons {
		t.Run(d.name, func(t *testing.T) {
			s, err := NewCronSchedule(d.schedule)

			require.NoError(t, err)
			require.NotNil(t, s)

			assert.Equal(t, d.schedule, s.Schedule())

			now := time.Now()
			c, _ := cron.ParseStandard(d.schedule)

			assert.Equal(t, c.Next(now), *s.Next(now))
		})
	}
}

func TestShouldCreateCronWithMaxDelay(t *testing.T) {
	for _, d := range validCrons {
		t.Run(d.name, func(t *testing.T) {
			md := 1 * time.Second
			s, err := NewCronScheduleWithMaxDelay(d.schedule, md)

			require.NoError(t, err)
			require.NotNil(t, s)

			if smd, ok := s.(JobScheduleWithMaxDelay); ok {
				assert.Equal(t, md, *smd.MaxDelay())
			} else {
				assert.Fail(t, "expected JobScheduleWithMaxDelay")
			}
		})
	}
}
