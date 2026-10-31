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
}
