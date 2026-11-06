import type { APIRequestContext } from "@playwright/test";
import { callRPC } from "./client";

export interface Campaign {
  id: string;
  name: string;
  members?: { id: string; name: string }[];
}

export namespace Campaign {
  export function randomName(): string {
    return "campaign-" + crypto.randomUUID();
  }

  /**
   * Creates a campaign as the user of the given token, that user becomes a member of it
   */
  export async function create(
    request: APIRequestContext,
    token: string,
    name: string = randomName(),
  ): Promise<Campaign> {
    const resp = await callRPC<{ campaign: Campaign }>(
      request,
      "campaigns.v1.CampaignService/Create",
      { name },
      token,
    );

    return resp.campaign;
  }

  /**
   * Adds the existing user with the given username to the campaign, as the user of the given token (the owner or an admin)
   */
  export async function addPlayer(
    request: APIRequestContext,
    token: string,
    campaignId: string,
    username: string,
  ): Promise<void> {
    await callRPC(
      request,
      "campaigns.v1.CampaignService/AddPlayer",
      { campaignId, username },
      token,
    );
  }
}
