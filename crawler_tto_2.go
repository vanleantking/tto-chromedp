package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// Constants and Configuration based on user's request and best practices.
const (
	PARTNER_TIKSHOP_HOME_URL = "https://ads.tiktok.com/creative/creator/explore?region=row&from_creative=login" // Replace with actual home URL
	TARGET_PAGE              = PARTNER_TIKSHOP_HOME_URL                                                         // The page where the search actually happens

	NAME_SEARCH_ELEM   = "//span[text()='Name search']"
	INPUT_SEARCH_ELEM  = `input:is([placeholder="Enter username or nickname"], [placeholder="Search names, products, hashtags, or keywords"])`
	SEARCH_BUTTON_ELEM = `button[data-testid="SearchKeyword-ExploreNameSearchInput-aVhwsM"]`

	SEARCH_RESULTS_BODY  = `div.virtualCardResults`
	FIRST_ROW_DATA_INDEX = "//div[@data-index='0']"
	GRID_ELEM            = ".gap-24"
	// Selector for the creator card section, handling dynamic testid
	SECTION_LOCATOR   = "section[data-testid='ExploreCreatorCard-index-coAQDf']"
	CREATOR_NAME_ELEM = ".text-black" // Assuming the creator's name is in a black text element
	NO_RESULTS_TEXT   = "No results found"

	DEFAULT_USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"
)

// KolData represents a single KOL entry from the input list
type KolData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// CollectedData holds the information captured from a matching network response.
type CollectedData struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
	// Add other necessary fields (e.g., body if required)
}

// --- New Tab Processing Logic (Conversion of _process_single_kol) ---

// processSingleKol performs the search, clicks the creator link, and captures network data in the new tab.
func processSingleKol(
	ctx context.Context,
	kolName string,
	urlPattern string,
) (string, []CollectedData, error) {

	// Slice to hold network data collected by the listener
	var collectedData []CollectedData
	var isEmailLocatorPresent bool // Variable to store the result of the locator check
	var isResultExists bool

	// Create a new context with a timeout for the KOL processing
	kolCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	// --- Step 1: Search and Wait for Results ---
	err := chromedp.Run(kolCtx,
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(NAME_SEARCH_ELEM, chromedp.BySearch),
		chromedp.Click(NAME_SEARCH_ELEM, chromedp.BySearch),
		chromedp.Sleep(5*time.Second),
		// --- START: LOCATOR EXISTENCE CHECK ---
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Check if the element exists in the DOM using JavaScript:
			// document.querySelector(selector) returns the element or null. Checking if it's not null gives a boolean.
			js := fmt.Sprintf("document.querySelector('%s') !== null", SEARCH_BUTTON_ELEM)

			// Execute the JavaScript and store the boolean result
			err := chromedp.Evaluate(js, &isEmailLocatorPresent).Do(ctx)
			if err != nil {
				// Log the error but continue, assuming the element is not present if the evaluation fails
				log.Printf("Error during locator existence check: %v", err)
				isEmailLocatorPresent = false
			}
			// Print the result as requested by the user
			log.Printf("Check for locator '%s': Exists = %t, %s, %s", SEARCH_BUTTON_ELEM, isEmailLocatorPresent)
			return nil
		}),
		// --- END: LOCATOR EXISTENCE CHECK ---

		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Waiting for search input to be visible...")
			return nil
		}),
		chromedp.WaitVisible(INPUT_SEARCH_ELEM, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Printf("Search input visible. Setting value to '%s'...", kolName)
			return nil
		}),
		// Use SendKeys for reliability with JS frameworks. It simulates user typing.
		chromedp.SendKeys(INPUT_SEARCH_ELEM, kolName, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Value set. Waiting before clicking search...")
			return nil
		}),
		chromedp.Sleep(2*time.Second), // A short pause can help ensure the UI has processed the input.
		chromedp.Click(SEARCH_BUTTON_ELEM, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Search button clicked. Waiting for results to load...")
			return nil
		}),
		chromedp.Sleep(15*time.Second),

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

		// --- END: LOCATOR EXISTENCE CHECK ---
		chromedp.WaitVisible(fmt.Sprintf("%s", SEARCH_RESULTS_BODY), chromedp.ByQuery),
		// chromedp.WaitVisible(fmt.Sprintf("%s %s %s %s", SEARCH_RESULTS_BODY, FIRST_ROW_DATA_INDEX, GRID_ELEM, SECTION_LOCATOR), chromedp.ByQuery),
		// --- START: LOCATOR EXISTENCE CHECK ---
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Check if the element exists in the DOM using JavaScript:
			// document.querySelector(selector) returns the element or null. Checking if it's not null gives a boolean.
			js := fmt.Sprintf("document.querySelector('%s') !== null", SEARCH_RESULTS_BODY)

			// Execute the JavaScript and store the boolean result
			err := chromedp.Evaluate(js, &isResultExists).Do(ctx)
			if err != nil {
				// Log the error but continue, assuming the element is not present if the evaluation fails
				log.Printf("Error during locator existence check: %v", err)
				isEmailLocatorPresent = false
			}
			// Print the result as requested by the user
			log.Printf("Check for locator '%s': Exists = %t, %s, %s", SEARCH_RESULTS_BODY, isResultExists)
			return nil
		}),
		// Scroll down to ensure the element is in the viewport (Playwright's window.scrollTo)
		// chromedp.Evaluate(`window.scrollTo(0, 500)`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("getting the text result...")
			return nil
		}),

		// Get the inner text of the entire first card for 'No results' check
		// chromedp.Text(fmt.Sprintf("%s div %s %s %s", SEARCH_RESULTS_BODY, FIRST_ROW_DATA_INDEX, GRID_ELEM, SECTION_LOCATOR), &firstCardContent, chromedp.ByQuery),

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
	clickTask := chromedp.Click(fmt.Sprintf("%s div div %s", SECTION_LOCATOR, CREATOR_NAME_ELEM), chromedp.ByQuery)

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

	var wg sync.WaitGroup
	// Start network listener on the new tab's context
	chromedp.ListenTarget(newTabCtx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			if strings.Contains(ev.Response.URL, urlPattern) {
				wg.Add(1)
				collectedData = append(collectedData, CollectedData{
					URL:    ev.Response.URL,
					Status: int(ev.Response.Status),
				})
				log.Printf("[NEW TAB RESPONSE] Captured %s (Status: %d)", ev.Response.URL, ev.Response.Status)
				wg.Done()
			}
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
	if err := chromedp.Run(newTabCtx, target.DetachFromTarget()); err != nil {
		log.Printf("Warning: Failed to detach/close new tab: %v", err)
	}
	log.Printf("New tab closed. Total captured responses: %d", len(collectedData))

	return kolName, collectedData, nil
}

// crawlerKols implements the main looping and state loading logic.
func crawlerKols(
	kolList []KolData,
	urlPattern string,
	statePath string,
	userAgent string,
	profileName string,
	headless bool,
) ([]KolData, error) {

	// 1. Initial Setup: Allocator Context (Browser Instance)
	opts := initChromedpOptions(profileName, headless, userAgent)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	// 3. Create Main Context (Tab) and Set Emulation
	mainTaskCtx, cancelMainTask := chromedp.NewContext(allocCtx)
	defer cancelMainTask()

	if err := chromedp.Run(mainTaskCtx,
		// 5. Set a common desktop viewport size
		chromedp.EmulateViewport(1920, 1080),
		// 6. Set locale/timezone for extra realism
		emulation.SetTimezoneOverride("Asia/Ho_Chi_Minh"), // <-- FIXED
		emulation.SetLocaleOverride(),                     // <-- FIXED
		// Inject cookies if they were loaded successfully
		// setCookieAction,
	); err != nil {
		return nil, fmt.Errorf("failed to set initial emulation/state: %w", err)
	}

	// 4. Navigate to the Target Page
	log.Printf("Navigating to protected page: %s", TARGET_PAGE)
	if err := chromedp.Run(mainTaskCtx,
		chromedp.Navigate(TARGET_PAGE),
		chromedp.WaitVisible(NAME_SEARCH_ELEM, chromedp.BySearch), // Wait for a key element to confirm load
		chromedp.Sleep(5*time.Second),                             // Deliberate pause
	); err != nil {
		return nil, fmt.Errorf("failed to navigate to target page %s: %w", TARGET_PAGE, err)
	}
	log.Println("Successfully navigated to the protected page.")

	// 5. Crawling Loop
	var finalCreators []KolData

	for _, kol := range kolList {
		log.Printf("\nProcessing KOL: ID=%s, Username=%s", kol.ID, kol.Username)

		// FIX: Pass the main tab's context (mainTaskCtx) to perform actions on the page.
		finalKolName, collectedData, err := processSingleKol(mainTaskCtx, kol.Username, urlPattern)

		if err != nil {
			log.Printf("Error processing KOL %s: %v", finalKolName, err)
			continue
		}

		log.Printf("Processed %s. Captured %d data points.", finalKolName, len(collectedData))

		// NOTE: Placeholder for your Python's self._parse_user_data logic.
		// You would typically process 'collectedData' here to create a final structure.
		// For now, we'll just track the successful names.
		if len(collectedData) > 0 {
			finalCreators = append(finalCreators, kol)
		}
	}

	// Final 5-second pause and browser close is handled by the defer cancelAlloc()
	log.Println("Finished processing all KOLs. Browser will close.")

	return finalCreators, nil
}

func main() {
	// --- Input Configuration ---
	// NOTE: This array of KOLs replaces the Python input list
	kolsToCrawl := []KolData{
		{ID: "1001", Username: "fayemabini"},
	}

	// loginURL := PARTNER_TIKTOKSHOP_LOGIN_URL
	// username := "van.le@brancherx.com" // Placeholder
	// password := "VLantking2013!"       // Placeholder
	statePath := "tiktokshop_state_go.json"
	urlPattern := "CreativeOne/MatchPack/MGetCreatorsCard" // Replace with the actual API endpoint pattern
	userAgent := DEFAULT_USER_AGENT
	profileName := "tto"

	// 2. Start the main crawling loop using the saved state.
	log.Println("\n--- Starting KOL Crawling Process ---")

	crawledData, err := crawlerKols(kolsToCrawl, urlPattern, statePath, userAgent, profileName, false)

	if err != nil {
		log.Fatalf("Fatal: Crawler failed: %v", err)
	}

	log.Printf("\n--- Final Summary ---")
	log.Printf("Total KOLs processed: %d", len(kolsToCrawl))
	log.Printf("Total creators successfully crawled (data captured): %d", len(crawledData))

	log.Println("Chromedp script finished successfully.")
}

// initChromedpOptions sets up the allocator options with anti-detection flags and user data.
func initChromedpOptions(profileName string, headless bool, userAgent string) []chromedp.ExecAllocatorOption {
	profilePath := filepath.Join(".", "profiles", profileName)
	fmt.Println("profilePath---------, ", profilePath)

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(profilePath),
		chromedp.Flag("headless", false),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)
	return opts
}
