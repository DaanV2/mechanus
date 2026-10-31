import { createClient, type Client, type Transport } from '@connectrpc/connect';
import { CampaignService } from '../../proto/campaigns/v1/campaign_pb';

export function createCampaignClient(transport: Transport): Client<typeof CampaignService> {
  return createClient(CampaignService, transport);
}
