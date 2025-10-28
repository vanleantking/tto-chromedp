package main

import (
	"context"
	"fmt"
	"log"
	"os" // Required for profile path handling
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/emulation" // <--- ADDED EMULATION IMPORT
	"github.com/chromedp/cdproto/network"

	// <-- FIXED: Renamed import to cdpnetwork
	"github.com/chromedp/chromedp"
)

// Constants and Configuration based on user's request and best practices.
const (
	// Placeholder for the login page and expected dashboard URL
	// Updated URLs based on the user's local environment testing
	PARTNER_TIKTOKSHOP_LOGIN_URL = "https://ads.tiktok.com/i18n/login?redirect=https%3A%2F%2Fads.tiktok.com%2Fcreative%2Flogin%3Fredirect%3Dhttps%253A%252F%252Fads.tiktok.com%252Fcreative%252Fcreator%252Fexplore%253Fregion%253Drow&_source_=tiktok-one" // Replace with actual login URL
	PARTNER_TIKSHOP_HOME_URL     = "https://ads.tiktok.com/creative/creator/explore?region=row&from_creative=login"                                                                                                                                         // Replace with actual home URL

	EMAIL_SELECTOR    = `input[placeholder="Enter your email address"]`
	PASSWORD_SELECTOR = `input[placeholder="Enter your password"]`
	LOGIN_BUTTON      = `.login-btn`

	DEFAULT_USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"
)

// UserState holds the necessary session data (cookies and local storage)
type UserState struct {
	Cookies []*network.Cookie `json:"cookies"`
	// LocalStorage is more complex to extract and is omitted for simplicity,
	// but can be added via custom JavaScript execution actions.
}

// simulateLogin logs into the platform and saves the session state (cookies).
func simulateLogin(
	loginURL string,
	username string,
	password string,
	statePath string,
	userAgent string,
	profileName string, // Added profileName to function signature
	headless bool,
	proxy string, // Note: Proxy configuration is more complex in chromedp and often requires external tools or specific transport settings.
) error {
	opts := initOPTTTS(profileName, headless, userAgent)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create browser context with a timeout for the entire login process
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Apply context settings for realism (locale, timezone, viewport)
	// These are typically set using the Emulation domain in CDP.
	if err := chromedp.Run(taskCtx,
		// 5. Set a common desktop viewport size
		chromedp.EmulateViewport(1920, 1080),
		// 6. Set locale/timezone for extra realism
		emulation.SetTimezoneOverride("Asia/Ho_Chi_Minh"), // <-- FIXED
		emulation.SetLocaleOverride(),                     // <-- FIXED
	); err != nil {
		return fmt.Errorf("failed to set emulation settings: %w", err)
	}

	// Create a new context with a separate timeout for navigation and interaction tasks
	interactionCtx, cancelInteraction := context.WithTimeout(taskCtx, 120*time.Second) // 120 second total timeout
	defer cancelInteraction()

	// --- 2. Define Chromedp Tasks (Login Flow) ---

	var currentURL string // Variable to store the current URL for verification
	var screenshotData []byte
	var isEmailLocatorPresent bool // Variable to store the result of the locator check

	err := chromedp.Run(interactionCtx,
		// Navigate to login page
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Printf("Navigating to %s", loginURL)
			return nil
			// return network.SetExtraHTTPHeaders(network.Headers{
			// 	"Accept-Language": "vi-VN,vi;q=0.9,en-US;q=0.8,en;q=0.7", // Adjusting Accept-Language based on vi-VN locale
			// }).Do(ctx)
		}),
		chromedp.Navigate(loginURL),

		// **IMPROVED WAIT:** Wait for the body and initial page load to ensure stability
		chromedp.WaitReady("body", chromedp.ByQuery),

		// Wait for login form visibility
		chromedp.WaitVisible(EMAIL_SELECTOR, chromedp.ByQuery),
		// --- START: LOCATOR EXISTENCE CHECK ---
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Check if the element exists in the DOM using JavaScript:
			// document.querySelector(selector) returns the element or null. Checking if it's not null gives a boolean.
			js := fmt.Sprintf("document.querySelector('%s') !== null", EMAIL_SELECTOR)

			// Execute the JavaScript and store the boolean result
			err := chromedp.Evaluate(js, &isEmailLocatorPresent).Do(ctx)
			if err != nil {
				// Log the error but continue, assuming the element is not present if the evaluation fails
				log.Printf("Error during locator existence check: %v", err)
				isEmailLocatorPresent = false
			}
			// Print the result as requested by the user
			log.Printf("Check for locator '%s': Exists = %t, %s, %s", EMAIL_SELECTOR, isEmailLocatorPresent, username, password)
			return nil
		}),
		// --- END: LOCATOR EXISTENCE CHECK ---

		// Fill the form with added delays
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Filling login form...")
			return chromedp.Tasks{
				// Use Clear and SetValue for reliability (more direct than Click + SendKeys)
				chromedp.Clear(EMAIL_SELECTOR, chromedp.ByQuery),
				chromedp.SetValue(EMAIL_SELECTOR, username, chromedp.ByQuery),
				chromedp.Sleep(1 * time.Second), // Wait briefly

				chromedp.Clear(PASSWORD_SELECTOR, chromedp.ByQuery),
				chromedp.SetValue(PASSWORD_SELECTOR, password, chromedp.ByQuery),
				chromedp.Sleep(2 * time.Second), // Wait before clicking submit
			}.Do(ctx)
		}),

		// Click the login button
		chromedp.Click(LOGIN_BUTTON, chromedp.ByQuery),

		// **RESTORED CRITICAL STEP:** Wait for the URL to change to the home page
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Waiting for navigation to the home page after login...")
			// Loop until the URL changes to the expected home URL or timeout
			for i := 0; i < 60; i++ { // Check for up to 60 seconds (half of the context timeout)
				time.Sleep(1 * time.Second)
				var url string
				if err := chromedp.Location(&url).Do(ctx); err != nil {
					return err
				}
				// Check for successful redirection
				if url == PARTNER_TIKSHOP_HOME_URL {
					log.Printf("Successfully redirected to: %s", url)
					currentURL = url
					return nil
				}
				// Check if we are still on the login page or an intermediate page
				if url != loginURL {
					currentURL = url
				}
			}
			return fmt.Errorf("timed out waiting for redirection to %s. Current URL: %s", PARTNER_TIKSHOP_HOME_URL, currentURL)
		}),

		// Take a screenshot
		chromedp.CaptureScreenshot(&screenshotData),

		// --- RESTORED: Save State (Extract Cookies) ---
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Save screenshot
			if len(screenshotData) > 0 {
				screenshotFile := "dashboard_after_login_go.png"
				if err := os.WriteFile(screenshotFile, screenshotData, 0644); err != nil {
					log.Printf("Warning: Failed to save screenshot: %v", err)
				} else {
					log.Printf("Screenshot of dashboard saved to '%s'", screenshotFile)
				}
			}

			log.Printf("Saving session state (cookies) to %s...", statePath)
			log.Printf("Session state saved to %s", statePath)

			// Keep browser open for a bit
			log.Println("Login and state saving complete. Closing browser in 5 seconds...")
			time.Sleep(500 * time.Second)
			return nil
		}),
	)

	if err != nil {
		// Log the last known URL on error
		if currentURL != "" {
			log.Printf("An error occurred. Last known URL: %s", currentURL)
		} else {
			log.Printf("An error occurred: %s", err.Error())
		}
		return err
	}

	return nil
}

func main() {
	// Dummy variables for demonstration (replace with actual configuration)
	// loginURL := PARTNER_TIKTOKSHOP_LOGIN_URL
	ttoHomepage := PARTNER_TIKSHOP_HOME_URL
	// username := "van.le@brancherx.com" // Placeholder
	// password := "VLantking2013!"       // Placeholder
	// statePath := "tiktokshop_state_go.json"
	userAgent := DEFAULT_USER_AGENT

	// err := simulateLogin(
	// 	loginURL,
	// 	username,
	// 	password,
	// 	statePath,
	// 	userAgent,
	// 	"tto",
	// 	false, // headless
	// 	"",    // proxy
	// )

	// if err != nil {
	// 	log.Fatalf("Login and state saving failed: %v", err)
	// 	return
	// }

	if err := visitHomePage(
		ttoHomepage,
		userAgent,
		"tto",
		false, // headless
		"",    // proxy
	); err != nil {
		log.Fatalf("Login and state saving failed: %v", err)
	}

	log.Println("Chromedp script finished successfully.")
}

func initOPTTTS(profileName string, headless bool, userAgent string) []chromedp.ExecAllocatorOption {
	// --- 1. Set up Chromedp Context and Options ---
	profilePath := filepath.Join("./profiles", profileName)

	// Define browser options (based on BROWSER_ARGS from the Python script for realism)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		// Anti-detection flag equivalent:
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),

		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("credentials_enable_service", false),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),

		// Set realistic user agent
		chromedp.UserAgent(userAgent),
		chromedp.UserDataDir(profilePath),
		chromedp.Flag("profile-directory", profileName),
	)

	if !headless {
		opts = append(opts, chromedp.Flag("headless", false))
	} else {
		// Use the new headless mode
		opts = append(opts, chromedp.Headless)
	}

	var tmp = make([]chromedp.ExecAllocatorOption, len(opts))
	copy(tmp, opts)

	return tmp
}
