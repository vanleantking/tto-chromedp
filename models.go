package main

// TTOCreatorResponse represents the structure of the response from the TikTok Creator API.
type TTOCreatorResponse struct {
	BaseResp struct {
		StatusCode    int    `json:"StatusCode"`
		StatusMessage string `json:"StatusMessage"`
	} `json:"baseResp"`
	Creators []struct {
		AioCreatorID  string `json:"aioCreatorID"`
		ContentLabels []struct {
			LabelID   string `json:"labelID"`
			LabelName string `json:"labelName"`
		} `json:"contentLabels"`
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
		IsCarveOut  bool `json:"isCarveOut"`
		PriceIndex  int  `json:"priceIndex"`
		RecentItems []struct {
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
		} `json:"recentItems"`
		RiskInfo struct {
			CreatorID string `json:"creatorID"`
		} `json:"riskInfo"`
		StatisticData struct {
			Algo struct {
				ContentLanguage []string `json:"contentLanguage"`
			} `json:"algo"`
			FollowerCountHistory struct {
				FollowerCount []struct {
					Count int    `json:"count"`
					Date  string `json:"date"`
				} `json:"followerCount"`
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
				Age []struct {
					AgeInterval string  `json:"ageInterval"`
					Ratio       float64 `json:"ratio"`
				} `json:"age"`
				DeviceBrand []struct {
					DeviceBrand string  `json:"deviceBrand"`
					Ratio       float64 `json:"ratio"`
				} `json:"deviceBrand"`
				Gender []struct {
					Gender string  `json:"gender"`
					Ratio  float64 `json:"ratio"`
				} `json:"gender"`
				Region []struct {
					Country string  `json:"country"`
					Ratio   float64 `json:"ratio"`
				} `json:"region"`
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
				RecentVideos []struct {
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
				} `json:"recentVideos"`
			} `json:"videoPerformance"`
		} `json:"statisticData"`
		TtUID string `json:"ttUID"`
	} `json:"creators"`
}
