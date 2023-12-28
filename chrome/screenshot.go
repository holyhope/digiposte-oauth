package chrome

import (
	"context"
	"image/jpeg"

	"github.com/chromedp/chromedp"
)

func (c *chromeLogin) ScreenshotIfNeeded(ctx context.Context, errPtr *error) {
	if c.screenShortOnError && *errPtr != nil {
		*errPtr = c.wrapWithScreenshot(ctx, *errPtr)
	}
}

func (c *chromeLogin) wrapWithScreenshot(ctx context.Context, rootErr error) error {
	var imageData []byte

	if err := chromedp.Run(ctx, chromedp.FullScreenshot(&imageData, jpeg.DefaultQuality)); err != nil {
		errorLogger(ctx).Printf("Failed to take screenshot: %v\n", err)

		return rootErr
	}

	infoLogger(ctx).Println("Screenshot taken")

	return &WithScreenshotError{
		Err:        rootErr,
		Screenshot: imageData,
	}
}
