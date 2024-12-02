package campaigns

import "github.com/DaanV2/mechanus/server/pkg/config"

var (
	CampaignConfig = config.New("campaigns")
	CampaignFolder = CampaignConfig.String("campaigns.folder", config.StorageFolder("campaigns"), "The folder where campaign data is stored")
)
