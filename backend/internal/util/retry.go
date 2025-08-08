package util

import (
	"log"
	"strings"
	"time"
)

// RetryOnLock retries the given function if it fails with a database lock error
func RetryOnLock(operation func() error) error {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		// Check if the error is a database lock error
		if strings.Contains(err.Error(), "database is locked") {
			// Exponential backoff: 100ms, 200ms, 400ms
			delay := baseDelay * time.Duration(1<<i)
			log.Printf("Database locked, retrying in %v...", delay)
			time.Sleep(delay)
			continue
		}

		// If it's not a lock error, return immediately
		return err
	}

	// If we've exhausted all retries, return the last error
	return err
}

// RetryOnLockWithResult retries the given function if it fails with a database lock error
// and returns the result along with any error
func RetryOnLockWithResult[T any](operation func() (T, error)) (T, error) {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	var result T
	var err error

	for i := 0; i < maxRetries; i++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}

		// Check if the error is a database lock error
		if strings.Contains(err.Error(), "database is locked") {
			// Exponential backoff: 100ms, 200ms, 400ms
			delay := baseDelay * time.Duration(1<<i)
			log.Printf("Database locked, retrying in %v...", delay)
			time.Sleep(delay)
			continue
		}

		// If it's not a lock error, return immediately
		return result, err
	}

	// If we've exhausted all retries, return the last result and error
	return result, err
}
