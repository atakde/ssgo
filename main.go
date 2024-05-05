package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

type SSUrl struct {
	URL      string
	Selector string
	Prefix   string
}

func main() {
	fmt.Println("Starting...")

	urls := []SSUrl{
		{URL: "https://keeplearning.dev/", Selector: `div.hero`, Prefix: "k"},
	}

	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	var wg sync.WaitGroup

	for _, url := range urls {
		// Increment the wait group counter for each URL
		wg.Add(1)

		// Execute tasks concurrently inside a goroutine
		go func(u SSUrl) {
			defer wg.Done()
			log.Printf("Capturing screenshot for URL: %s", u.URL)
			captureScreenshot(ctx, u)
		}(url)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("All screenshots captured.")

}

func captureScreenshot(ctx context.Context, u SSUrl) {
	var buf []byte

	if err := chromedp.Run(ctx, elementScreenshot(u.URL, u.Selector, &buf)); err != nil {
		log.Printf("Error capturing screenshot for URL %s: %v", u.URL, err)
		return
	}

	filename := fmt.Sprintf("%s-screenshot-%d.png", u.Prefix, time.Now().Unix())
	if err := os.WriteFile(filename, buf, 0o644); err != nil {
		log.Printf("Error saving screenshot for URL %s: %v", u.URL, err)
		return
	}

	fmt.Printf("Screenshot saved for URL: %s\n", u.URL)
}

func elementScreenshot(urlstr string, selector string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var width, height int64
			if err := chromedp.Evaluate(`document.body.scrollWidth`, &width).Do(ctx); err != nil {
				return err
			}

			if err := chromedp.Evaluate(`document.body.scrollHeight`, &height).Do(ctx); err != nil {
				return err
			}

			log.Println("width:", width)
			log.Println("height:", height)

			if width == 0 || height == 0 {
				return fmt.Errorf("document body width or height is zero")
			}

			err := emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			return nil
		}),
		chromedp.Screenshot(selector, res, chromedp.NodeVisible, chromedp.ByQuery),
	}
}
