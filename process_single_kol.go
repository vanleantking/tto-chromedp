package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

func processSingleKol2(
	ctx context.Context,
	kolName string,
	urlPattern string,
) (string, []CollectedData, error) {

	// Slice to hold network data collected by the listener
	var collectedData []CollectedData

	// Create a new context with a timeout for the KOL processing
	kolCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	// --- Step 1: Search and Wait for Results ---
	err := chromedp.Run(kolCtx,
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(NAME_SEARCH_ELEM, chromedp.BySearch),
		chromedp.Click(NAME_SEARCH_ELEM, chromedp.BySearch),
		chromedp.Sleep(1*time.Second),

		chromedp.WaitVisible(INPUT_SEARCH_ELEM, chromedp.ByQuery),
		chromedp.SetValue(INPUT_SEARCH_ELEM, kolName, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		// ACTION MODIFICATION: Click on the specific Search Button
		chromedp.Click(SEARCH_BUTTON_ELEM, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		// Wait for search results table body
		chromedp.WaitVisible(SEARCH_RESULTS_BODY, chromedp.ByQuery),
	)
	if err != nil {
		return kolName, nil, fmt.Errorf("search failed: %w", err)
	}
	log.Println("Search results table visible. Checking content...")

	// --- Step 2: Validate Search Result and Prepare Click Target ---
	var firstCardContent string
	var creatorNameText string

	// Get content and creator name from the first result card
	err = chromedp.Run(kolCtx,
		// Get the full HTML content of the first result card
		chromedp.WaitVisible(fmt.Sprintf("%s", SEARCH_RESULTS_BODY), chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		// Get the specific creator name element text
		chromedp.Text(fmt.Sprintf("%s div div %s", SECTION_LOCATOR, CREATOR_NAME_ELEM), &creatorNameText, chromedp.ByQuery),
	)
	if err != nil {
		return kolName, nil, fmt.Errorf("failed to retrieve search result content: %w", err)
	}

	if strings.Contains(firstCardContent, NO_RESULTS_TEXT) || strings.TrimSpace(creatorNameText) != kolName {
		log.Printf("No results found or name mismatch: Expected '%s', Found '%s'", kolName, creatorNameText)
		return kolName, collectedData, nil
	}
	log.Printf("Found matching creator: %s. Proceeding to click.", creatorNameText)

	// --- Step 3: Click and Capture New Tab ---

	// The listener channel for the new target ID
	targetCh := make(chan target.ID, 1)

	// Start listener for new targets (new tabs)
	// This listener must be set up *before* the click.
	chromedp.ListenTarget(kolCtx, func(ev interface{}) {
		if ev, ok := ev.(*target.EventTargetCreated); ok {
			// Filter out targets that aren't of type 'page' (e.g., workers, iframes)
			if ev.TargetInfo.Type == "page" {
				targetCh <- ev.TargetInfo.TargetID
			}
		}
	})

	// The click task runs concurrently with the listener
	clickTask := chromedp.Click(fmt.Sprintf("%s %s %s %s %s", SEARCH_RESULTS_BODY, FIRST_ROW_DATA_INDEX, GRID_ELEM, SECTION_LOCATOR, CREATOR_NAME_ELEM), chromedp.ByQuery)

	// Perform the click, which triggers the new target event
	if err := chromedp.Run(kolCtx, clickTask); err != nil {
		return kolName, nil, fmt.Errorf("failed to click creator link: %w", err)
	}

	var newTargetID target.ID
	select {
	case newTargetID = <-targetCh:
		log.Printf("SUCCESS: Captured new target ID: %s", newTargetID)
	case <-time.After(15 * time.Second):
		return kolName, nil, fmt.Errorf("timed out waiting for new tab target event")
	}

	// --- Step 4: Create New Context and Process Tab ---

	// Create a new context attached to the new tab.
	// NOTE: We pass the main Allocator Context (ctx) so the new tab is part of the same browser instance.
	newTabCtx, cancelNewTab := chromedp.NewContext(ctx, chromedp.WithTargetID(newTargetID))
	defer cancelNewTab()

	// FIX 1: Add a delay and run to ensure the target session is fully attached and the context is valid.
	if err := chromedp.Run(newTabCtx, chromedp.Sleep(500*time.Millisecond)); err != nil {
		return kolName, nil, fmt.Errorf("failed to stabilize new tab context: %w", err)
	}

	// FIX 2: Now enable network events.
	if err := network.Enable().Do(newTabCtx); err != nil {
		return kolName, nil, fmt.Errorf("failed to enable network events on new tab: %w", err)
	}

	var wg sync.WaitGroup
	// Start network listener on the new tab's context
	chromedp.ListenTarget(newTabCtx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			if !strings.Contains(ev.Response.URL, urlPattern) {
				return
			}

			wg.Add(1)
			// This goroutine is necessary because GetResponseBody blocks,
			// and we don't want to block the event listener.
			go func() {
				defer wg.Done()
				// We must use the context of the new tab for GetResponseBody
				body, err := network.GetResponseBody(ev.RequestID).Do(newTabCtx)

				if err != nil {
					// FIX: Simplified the GetResponseBody call. If this error persists, it means the response
					// was already closed or the browser is blocking it, but the error message should change.
					log.Printf("Error getting response body for %s: %v", ev.Response.URL, err)
					return
				}

				// The JSON unmarshalling logic is correct but needs the TTOCreatorResponse struct
				var ttoResp TTOCreatorResponse
				if err := json.Unmarshal(body, &ttoResp); err != nil {
					log.Printf("Error unmarshalling response for %s: %v", ev.Response.URL, err)
					return
				}

				collectedData = append(collectedData, CollectedData{URL: ev.Response.URL, Status: int(ev.Response.Status), Body: &ttoResp})
				log.Printf("[NEW TAB RESPONSE] Captured and unmarshalled %s (Status: %d)", ev.Response.URL, ev.Response.Status)
			}()
		}
	})

	// Run actions on the new tab
	if err := chromedp.Run(newTabCtx,
		chromedp.EmulateViewport(1920, 1080),
		// Wait for the page content to fully load
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // Wait for async data load (Playwright's 100000ms equivalent, but reduced)
	); err != nil {
		return kolName, nil, fmt.Errorf("failed to load/process new tab: %w", err)
	}

	// Ensure all listeners have finished processing
	wg.Wait()

	// Close the new tab's target
	// FIX 3: Use the original context (kolCtx) to detach the target, which is generally more reliable.
	if err := target.DetachFromTarget().Do(kolCtx); err != nil {
		log.Printf("Warning: Failed to detach/close new tab: %v", err)
	}
	log.Printf("New tab closed. Total captured responses: %d", len(collectedData))

	return kolName, collectedData, nil
}
