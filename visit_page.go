package main

import (
	"context"
	"fmt"
	"log" // Required for profile path handling
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/emulation" // <--- ADDED EMULATION IMPORT
	"github.com/chromedp/cdproto/network"

	// <-- FIXED: Renamed import to cdpnetwork
	"github.com/chromedp/chromedp"
)

// Constants and Configuration based on user's request and best practices.
// visitHomePage logs into the platform and saves the session state (cookies).
func visitHomePage(
	loginURL string,
	userAgent string,
	profileName string, // Added profileName to function signature
	headless bool,
	proxy string, // Note: Proxy configuration is more complex in chromedp and often requires external tools or specific transport settings.
) error {
	opts := initChromedpOptions(profileName, headless, userAgent)

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
	interactionCtx, cancelInteraction := context.WithTimeout(taskCtx, 600*time.Second) // 120 second total timeout
	defer cancelInteraction()

	// --- 2. Define Chromedp Tasks (Login Flow) ---

	var currentURL string // Variable to store the current URL for verification

	err := chromedp.Run(interactionCtx,
		// Navigate to login page
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Printf("Navigating to %s", loginURL)
			return network.SetExtraHTTPHeaders(network.Headers{
				"Accept-Language": "vi-VN,vi;q=0.9,en-US;q=0.8,en;q=0.7", // Adjusting Accept-Language based on vi-VN locale
			}).Do(ctx)
		}),
		chromedp.Navigate(loginURL),

		// **IMPROVED WAIT:** Wait for the body and initial page load to ensure stability
		// chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(600*time.Second),
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

// initChromedpOptions sets up the allocator options with anti-detection flags and user data.
func initChromedpOptions(profileName string, headless bool, userAgent string) []chromedp.ExecAllocatorOption {
	profilePath := filepath.Join("./profiles/", profileName)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		// Anti-detection flags
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false), // Kept this as false

		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("credentials_enable_service", false),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),

		// User data settings
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
	return opts
}
