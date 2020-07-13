package util

import "time"

// Await runs the provided `test` function once every 100ms and returns as soon
// as either:
//
// * The function returns nil, indicating success, or
// * The function returns an error and we've exceeded the provided timeout.
func Await(test func() error, timeoutDuration time.Duration) error {
	timeout := time.After(timeoutDuration)

	for {
		err := test()

		if err == nil {
			return nil
		}

		select {
		case <-timeout:
			return err
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
