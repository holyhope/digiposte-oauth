package chrome

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/chromedp"
)

type Screen interface {
	fmt.Stringer
	CurrentPageMatches(ctx context.Context) bool
	ShouldWaitForResponse() bool
	chromedp.Action
}

type Screens struct {
	screens          []Screen
	refreshFrequency time.Duration

	succeeded atomic.Bool
}

func (s *Screens) Resolve(ctx context.Context) {
	var waitGroup sync.WaitGroup

	infoLgr, errorLgr := infoLogger(ctx), errorLogger(ctx)
	infoPrefix, errorPrefix := infoLgr.Prefix(), errorLgr.Prefix()

	defer infoLgr.Println("Stopped all resolvers")

	for _, screen := range s.screens {
		infoLgr.SetPrefix(fmt.Sprintf("%s[%v] ", infoPrefix, screen))
		errorLgr.SetPrefix(fmt.Sprintf("%s[%v] ", errorPrefix, screen))

		ctx := withErrorLogger(withInfoLogger(ctx, infoLgr), errorLgr)

		waitGroup.Add(1)

		go func(ctx context.Context, screen Screen) {
			defer waitGroup.Done()

			s.resolve(ctx, screen)
		}(ctx, screen)
	}

	waitGroup.Wait()
}

func (s *Screens) resolve(ctx context.Context, screen Screen) {
	refreshFrequency := time.NewTicker(s.refreshFrequency)
	defer refreshFrequency.Stop()

	infoLogger(ctx).Println("Started resolver...")
	defer infoLogger(ctx).Println("Stopped resolver")

	for !s.succeeded.Load() {
		select {
		case <-ctx.Done():
			return

		case <-refreshFrequency.C:
			func(ctx context.Context) {
				ctx, cancel := context.WithTimeout(ctx, s.refreshFrequency)
				defer cancel()

				if screen.CurrentPageMatches(ctx) {
					infoLogger(ctx).Println("Resolving screen...")

					if err := resolve(ctx, screen); err != nil {
						if !errors.Is(err, context.DeadlineExceeded) {
							errorLogger(ctx).Printf("Failed to run in chrome: %v\n", err)
						} else {
							infoLogger(ctx).Printf("Screen failed: %v\n", err)
						}

						return
					}

					infoLogger(ctx).Println("Screen passed")
				}
			}(ctx)
		}
	}
}

func (s *Screens) Succeeded() bool {
	return s.succeeded.Load()
}

func resolve(ctx context.Context, screen Screen) error {
	if screen.ShouldWaitForResponse() {
		resp, err := chromedp.RunResponse(ctx, screen)
		if err != nil {
			return fmt.Errorf("run response: %w", err)
		}

		if resp != nil && resp.Status >= 400 {
			return fmt.Errorf("response: %w", &HTTPError{
				Status:     resp.Status,
				StatusText: resp.StatusText,
			})
		}

		return nil
	}

	if err := chromedp.Run(ctx, screen); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}
