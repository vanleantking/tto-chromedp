package main

import (
	"context"
	"fmt"
	"log" // Required for profile path handling
	"path/filepath"
	"time"

	// <-- FIXED: Renamed import to cdpnetwork
	"github.com/chromedp/chromedp"
)

// Constants and Configuration based on user's request and best practices.
// visitHomePage logs into the platform and saves the session state (cookies).
func visitHomePage(
	loginURL string,
	profileName string, // Added profileName to function signature
	headless bool,
	proxy string, // Note: Proxy configuration is more complex in chromedp and often requires external tools or specific transport settings.
) error {
	opts := initChromedpOptions(profileName, headless)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	if err := chromedp.Run(taskCtx, chromedp.EmulateViewport(1920, 1080)); err != nil {
		return fmt.Errorf("failed to set emulation settings: %w", err)
	}

	interactionCtx, cancelInteraction := context.WithTimeout(taskCtx, 600*time.Second) // 120 second total timeout
	defer cancelInteraction()

	var currentURL string
	err := chromedp.Run(interactionCtx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(600*time.Second),
	)

	if err != nil {
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
func initChromedpOptions(profileName string, headless bool) []chromedp.ExecAllocatorOption {
	profilePath := filepath.Join(".", "profiles", profileName)

	fmt.Printf("profilePath %+v\n", profilePath)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserDataDir(profilePath),
		chromedp.Flag("headless", false),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	}

	if !headless {
		opts = append(opts, chromedp.Flag("headless", false))
	}

	return opts

	//opts := append(chromedp.DefaultExecAllocatorOptions[:],
	//	chromedp.NoFirstRun,
	//	chromedp.NoDefaultBrowserCheck,
	//	// Anti-detection flags
	//	chromedp.Flag("disable-blink-features", "AutomationControlled"),
	//	chromedp.Flag("enable-automation", false),
	//	chromedp.Flag("disable-extensions", false), // Kept this as false
	//	chromedp.Flag("disable-infobars", true),
	//	chromedp.Flag("disable-gpu", true),
	//	chromedp.Flag("allow-running-insecure-content", true),
	//	chromedp.Flag("disable-notifications", true),
	//	chromedp.Flag("disable-translate", true),
	//	chromedp.Flag("password-store", "basic"),
	//	chromedp.Flag("credentials_enable_service", false),
	//	chromedp.Flag("disable-default-apps", true),
	//	chromedp.Flag("disable-dev-shm-usage", true),
	//
	//	// User data settings
	//	chromedp.UserAgent(userAgent),
	//	chromedp.UserDataDir(profilePath),
	//	chromedp.Flag("profile-directory", profileName),
	//)
	//
	//if !headless {
	//	opts = append(opts, chromedp.Flag("headless", false))
	//} else {
	//	// Use the new headless mode
	//	opts = append(opts, chromedp.Headless)
	//}
	//return opts
}
