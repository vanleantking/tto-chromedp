package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
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

// TTOCreatorResponse represents the structure of the response from the TikTok Creator API.
type TTOCreatorResponse struct {
	BaseResp struct {
		StatusCode    int    `json:"StatusCode"`
		StatusMessage string `json:"StatusMessage"`
	} `json:"baseResp"`
	Creators []struct {
		AioCreatorID   string         `json:"aioCreatorID"`
		ContentLabels  []ContentLabel `json:"contentLabels"`
		CreatorProfile struct {
			Price struct {
			} `json:"price"`
			SpokenLanguageList []string `json:"spokenLanguageList"`
		} `json:"creatorProfile"`
		CreatorTTInfo struct {
			AdCreativeClass int    `json:"adCreativeClass"`
			AioCreatorID    string `json:"aioCreatorID"`
			AvatarURI       string `json:"avatarURI"`
			AvatarURL       string `json:"avatarURL"`
			AvatarURLList   []struct {
				Format   string `json:"format"`
				ImageURL string `json:"imageUrl"`
			} `json:"avatarURLList"`
			Bio                 string `json:"bio"`
			BrandedContentClass int    `json:"brandedContentClass"`
			Categories          []int  `json:"categories"`
			CreditScore         struct {
				AioCreatorID    string `json:"aioCreatorID"`
				CurrentScore    int    `json:"currentScore"`
				CurrentTier     int    `json:"currentTier"`
				ScoreLowerLimit int    `json:"scoreLowerLimit"`
				ScoreUpperLimit int    `json:"scoreUpperLimit"`
			} `json:"creditScore"`
			DataVDCRegion   int    `json:"dataVDCRegion"`
			DisplayStatus   int    `json:"displayStatus"`
			FollowerCnt     int    `json:"followerCnt"`
			HandleName      string `json:"handleName"`
			IsBannedInTT    bool   `json:"isBannedInTT"`
			IsRegisteredAIO bool   `json:"isRegisteredAIO"`
			IsTest          bool   `json:"isTest"`
			LivingRegion    string `json:"livingRegion"`
			NickName        string `json:"nickName"`
			RiskInfo        struct {
				CreatorID string `json:"creatorID"`
			} `json:"riskInfo"`
			StoreRegion string `json:"storeRegion"`
			TtUID       string `json:"ttUID"`
		} `json:"creatorTTInfo"`
		CreatorType int `json:"creatorType"`
		CreditScore struct {
			AioCreatorID    string `json:"aioCreatorID"`
			CurrentScore    int    `json:"currentScore"`
			CurrentTier     int    `json:"currentTier"`
			ScoreLowerLimit int    `json:"scoreLowerLimit"`
			ScoreUpperLimit int    `json:"scoreUpperLimit"`
		} `json:"creditScore"`
		DisplayType int `json:"displayType"`
		EsData      struct {
			AppearOnSearchSetting bool  `json:"appearOnSearchSetting"`
			Categories            []int `json:"categories"`
			Price                 struct {
				Currency                    string `json:"currency"`
				RecommendRate100K           string `json:"recommendRate100k"`
				StartingRate100K            string `json:"startingRate100k"`
				StoreRegionCurrency         string `json:"storeRegionCurrency"`
				StoreRegionStartingRate100K string `json:"storeRegionStartingRate100k"`
			} `json:"price"`
			Status int `json:"status"`
		} `json:"esData"`
		IndustryLabels []struct {
			LabelID   string `json:"labelID"`
			LabelName string `json:"labelName"`
		} `json:"industryLabels"`
		IsCarveOut  bool        `json:"isCarveOut"`
		PriceIndex  int         `json:"priceIndex"`
		RecentItems []VideoItem `json:"recentItems"`
		RiskInfo    struct {
			CreatorID string `json:"creatorID"`
		} `json:"riskInfo"`
		StatisticData struct {
			Algo struct {
				ContentLanguage []string `json:"contentLanguage"`
			} `json:"algo"`
			FollowerCountHistory struct {
				FollowerCount      []FollowerTrend `json:"followerCount"`
				FollowerGrowthRate []struct {
					Date string  `json:"date"`
					Rate float64 `json:"rate"`
				} `json:"followerGrowthRate"`
			} `json:"followerCountHistory"`
			FollowerDistriData struct {
				Active []struct {
					Active string  `json:"active"`
					Ratio  float64 `json:"ratio"`
				} `json:"active"`
				Age         []AgeDistri `json:"age"`
				DeviceBrand []struct {
					DeviceBrand string  `json:"deviceBrand"`
					Ratio       float64 `json:"ratio"`
				} `json:"deviceBrand"`
				Gender []GenderDistri `json:"gender"`
				Region []RegionDistri `json:"region"`
			} `json:"followerDistriData"`
			OverallPerformance struct {
				AvgSixSecondsViewsBenchMarkViews float64 `json:"avgSixSecondsViewsBenchMarkViews"`
				AvgSixSecondsViewsRate           float64 `json:"avgSixSecondsViewsRate"`
				AvgSixSecondsViewsRateRank       float64 `json:"avgSixSecondsViewsRateRank"`
				EngagementRate                   float64 `json:"engagementRate"`
				EngagementRateBenchMark          float64 `json:"engagementRateBenchMark"`
				EngagementRateRank               float64 `json:"engagementRateRank"`
				FollowerCount                    int     `json:"followerCount"`
				FollowerTier                     int     `json:"followerTier"`
				FollowersGrowthRate              float64 `json:"followersGrowthRate"`
				FollowersGrowthRateRank          float64 `json:"followersGrowthRateRank"`
				MedianBenchMarkViews             int     `json:"medianBenchMarkViews"`
				MedianViews                      int     `json:"medianViews"`
				MedianViewsRank                  float64 `json:"medianViewsRank"`
				VideoCompleteRate                float64 `json:"videoCompleteRate"`
				VideoCompleteRateRank            float64 `json:"videoCompleteRateRank"`
			} `json:"overallPerformance"`
			TtBasicInfo struct {
				AppLanguage []string `json:"appLanguage"`
			} `json:"ttBasicInfo"`
			VideoPerformance struct {
				PopularVideos []struct {
					Comment      int    `json:"comment"`
					CoverURL     string `json:"coverURL"`
					CoverURLList []struct {
						Format   string `json:"format"`
						ImageURL string `json:"imageUrl"`
					} `json:"coverURLList"`
					CreateTime       string `json:"createTime"`
					Heart            int    `json:"heart"`
					IsBoosted        bool   `json:"isBoosted"`
					IsSponsoredVideo bool   `json:"isSponsoredVideo"`
					ItemID           string `json:"itemID"`
					Share            int    `json:"share"`
					Title            string `json:"title"`
					VideoURL         string `json:"videoURL"`
					Views            string `json:"views"`
				} `json:"popularVideos"`
				RecentVideos []VideoItem `json:"recentVideos"`
			} `json:"videoPerformance"`
		} `json:"statisticData"`
		TtUID string `json:"ttUID"`
	} `json:"creators"`
}

type TTOUser struct {
	CategoryContent []ContentLabel  `json:"category_content"`
	AgeDistri       []AgeDistri     `json:"age_distri"`
	RegionDistri    []RegionDistri  `json:"region_distri"`
	GenderDistri    []GenderDistri  `json:"gender_distri"`
	FollowerTrend   []FollowerTrend `json:"follower_trend"`
	VideoViews      []VideoItem     `json:"video_views"`
}

type AgeDistri struct {
	AgeInterval string  `json:"ageInterval"`
	Ratio       float64 `json:"ratio"`
}

type GenderDistri struct {
	Gender string  `json:"gender"`
	Ratio  float64 `json:"ratio"`
}

type RegionDistri struct {
	Country string  `json:"country"`
	Ratio   float64 `json:"ratio"`
}

type FollowerTrend struct {
	Count int    `json:"count"`
	Date  string `json:"date"`
}

type ContentLabel struct {
	LabelID   string `json:"labelID"`
	LabelName string `json:"labelName"`
}

type VideoItem struct {
	Comment      int    `json:"comment"`
	CoverURL     string `json:"coverURL"`
	CoverURLList []struct {
		Format   string `json:"format"`
		ImageURL string `json:"imageUrl"`
	} `json:"coverURLList"`
	CreateTime       string `json:"createTime"`
	Heart            int    `json:"heart"`
	IsBoosted        bool   `json:"isBoosted"`
	IsSponsoredVideo bool   `json:"isSponsoredVideo"`
	ItemID           string `json:"itemID"`
	Share            int    `json:"share"`
	Title            string `json:"title"`
	VideoURL         string `json:"videoURL"`
	Views            string `json:"views"`
}

// CollectedData holds the information captured from a matching network response.
type CollectedData struct {
	URL    string              `json:"url"`
	Status int                 `json:"status"`
	Body   *TTOCreatorResponse `json:"body,omitempty"`
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

	// Create a new context with a timeout for the KOL processing
	kolCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	// --- Step 1: Search and Wait for Results ---
	err := chromedp.Run(kolCtx,
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(NAME_SEARCH_ELEM, chromedp.BySearch),
		chromedp.Click(NAME_SEARCH_ELEM, chromedp.BySearch),
		chromedp.Sleep(5*time.Second),
		chromedp.WaitVisible(INPUT_SEARCH_ELEM, chromedp.ByQuery),
		chromedp.Click(INPUT_SEARCH_ELEM, chromedp.BySearch),
		chromedp.Sleep(2*time.Second),
		// Use Clear to reliably empty the input field, avoiding stale element issues.
		chromedp.Clear(INPUT_SEARCH_ELEM, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		// Use SendKeys for reliability with JS frameworks. It simulates user typing.
		chromedp.SendKeys(INPUT_SEARCH_ELEM, kolName, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // A short pause can help ensure the UI has processed the input.
		chromedp.Click(SEARCH_BUTTON_ELEM, chromedp.ByQuery),
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
			if !strings.Contains(ev.Response.URL, urlPattern) {
				return
			}
			resp := ev.Response
			log.Printf("[NEW TAB RESPONSE] Captured %s (Status: %d): %s", resp.URL, resp.Status, ev.RequestID)

			wg.Add(1) // Increment counter before spawning goroutine
			// This goroutine is necessary because GetResponseBody blocks,
			// and we don't want to block the event listener.
			go func() {
				defer wg.Done()

				c := chromedp.FromContext(newTabCtx)
				body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(newTabCtx, c.Target))
				// body, err := network.GetResponseBody(ev.RequestID).Do(newTabCtx)
				if err != nil {
					log.Printf("Error getting response body for %s: %v", ev.Response.URL, err)
					return
				}
				fmt.Println("body response from url------------, ", resp.URL)

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
		// Refresh the page as requested
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Refreshing the page...")
			return nil
		}),
		chromedp.Reload(),
		chromedp.Sleep(5*time.Second), // Wait again after refresh
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
	kol KolData,
	urlPattern string,
	statePath string,
	userAgent string,
	profileName string,
	headless bool,
) ([]CollectedData, error) {

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

	log.Printf("\nProcessing KOL: ID=%s, Username=%s", kol.ID, kol.Username)

	// FIX: Pass the main tab's context (mainTaskCtx) to perform actions on the page.
	finalKolName, collectedData, err := processSingleKol(mainTaskCtx, kol.Username, urlPattern)

	if err != nil {
		log.Printf("Error processing KOL %s: %v", finalKolName, err)
		return nil, err
	}

	log.Printf("Processed %s. Captured %d data points.", finalKolName, len(collectedData))

	// NOTE: Placeholder for your Python's self._parse_user_data logic.
	// You would typically process 'collectedData' here to create a final structure.
	// For now, we'll just track the successful names.
	if len(collectedData) > 0 {
		log.Printf("Data captured for KOL %s.", kol.Username, len(collectedData))
	}

	// Final 5-second pause and browser close is handled by the defer cancelAlloc()
	log.Println("Finished processing all KOLs. Browser will close.")

	return collectedData, nil
}

func main() {
	// --- Input Configuration ---
	// NOTE: This array of KOLs replaces the Python input list
	kolsToCrawl := []KolData{
		{ID: "1001", Username: "fayemabini"},
		{ID: "1005", Username: "daipimenta"},
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

	for _, kol := range kolsToCrawl {

		crawledData, err := crawlerKols(kol, urlPattern, statePath, userAgent, profileName, false)

		if err != nil {
			log.Fatalf("Fatal: Crawler failed: %v", err)
			continue
		}
		log.Printf("Successfully crawled creator: ID=%s, Username=%s", kol.Username, len(crawledData))
		userInfo, isFull := parseUserData(crawledData)
		fmt.Println("------------user info parser, ", *userInfo, isFull)
	}

	log.Printf("\n--- Final Summary ---")

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

func parseUserData(collectedData []CollectedData) (*TTOUser, bool) {
	var categoryContent []ContentLabel
	var ageDistri []AgeDistri
	var regionDistri []RegionDistri
	var genderDistri []GenderDistri
	var followerTrend []FollowerTrend
	var videoViews []VideoItem

	var isFull = false

	for _, data := range collectedData {
		if data.Body != nil {
			dataResp := *data.Body
			// if not exist creator data, continue
			if len(dataResp.Creators) == 0 {
				continue
			}
			creatorData := dataResp.Creators[0]
			// Collect category labels
			if len(creatorData.ContentLabels) > 0 || creatorData.ContentLabels != nil {
				categoryContent = creatorData.ContentLabels
			}
			// Collect demographic distributions
			if len(creatorData.StatisticData.FollowerDistriData.Age) > 0 ||
				creatorData.StatisticData.FollowerDistriData.Age != nil {
				ageDistri = creatorData.StatisticData.FollowerDistriData.Age
			}

			// Collect region and gender distributions
			if len(creatorData.StatisticData.FollowerDistriData.Region) > 0 ||
				creatorData.StatisticData.FollowerDistriData.Region != nil {
				regionDistri = creatorData.StatisticData.FollowerDistriData.Region
			}
			if len(creatorData.StatisticData.FollowerDistriData.Gender) > 0 ||
				creatorData.StatisticData.FollowerDistriData.Gender != nil {
				genderDistri = creatorData.StatisticData.FollowerDistriData.Gender
			}
			// Collect follower trends
			if len(creatorData.StatisticData.FollowerCountHistory.FollowerCount) > 0 ||
				creatorData.StatisticData.FollowerCountHistory.FollowerCount != nil {
				followerTrend = creatorData.StatisticData.FollowerCountHistory.FollowerCount
			}
			// Collect video views
			if len(creatorData.StatisticData.VideoPerformance.RecentVideos) > 0 ||
				creatorData.StatisticData.VideoPerformance.RecentVideos != nil {
				videoViews = creatorData.StatisticData.VideoPerformance.RecentVideos
			}

			// If all data is collected, break
			if len(categoryContent) > 0 &&
				len(ageDistri) > 0 &&
				len(regionDistri) > 0 &&
				len(genderDistri) > 0 &&
				len(followerTrend) > 0 &&
				len(videoViews) > 0 {
				isFull = true
				break
			}
		}
	}
	if isFull {
		return &TTOUser{
			CategoryContent: categoryContent,
			AgeDistri:       ageDistri,
			RegionDistri:    regionDistri,
			GenderDistri:    genderDistri,
			FollowerTrend:   followerTrend,
			VideoViews:      videoViews,
		}, true
	}
	return nil, false
}
