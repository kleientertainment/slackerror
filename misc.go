package slack

import "time"

func maxDuration(l ...time.Duration) (max time.Duration) {
	for _, d := range l {
		if d > max {
			max = d
		}
	}
	return max
}
