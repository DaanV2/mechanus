<script lang="ts">
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import { createCampaignClient } from '$lib/stores/clients';
  import { timestampDate } from '@bufbuild/protobuf/wkt';
  import { Code, ConnectError } from '@connectrpc/connect';
  import { Heading, Hr, Listgroup, ListgroupItem } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import type { Campaign } from '../../../proto/campaigns/v1/campaign_pb';

  let campaign = $state<Campaign | undefined>(undefined);
  let errorObj = $state<MechanusError>(null);

  onMount(async () => {
    if (!userHandler.current.loggedin) {
      return goto('/users/login');
    }

    // The page is prerendered, so the id is read client side from the query string
    const id = new URLSearchParams(window.location.search).get('id') ?? '';
    if (id === '') {
      errorObj = 'No campaign selected';
      return;
    }

    try {
      const resp = await createCampaignClient().get({ id });
      campaign = resp.campaign;
    } catch (err) {
      if (err instanceof ConnectError && err.code === Code.NotFound) {
        errorObj = 'Campaign not found';
      } else {
        errorObj = err instanceof Error ? err : new Error('An unexpected error occurred.');
      }
    }
  });
</script>

<svelte:head>
  {#if campaign}
    <title>Campaign - {campaign.name}</title>
  {:else}
    <title>Campaign</title>
  {/if}
</svelte:head>

<NavBar />

<div class="px-5 py-3">
  <a href="/campaigns" class="text-sm hover:underline">&larr; Back to campaigns</a>
</div>

{#if campaign}
  <Heading tag="h2" class="px-5 pb-3">{campaign.name}</Heading>
  <Hr />

  <div class="max-w-md space-y-4 px-5">
    {#if campaign.createdAt}
      <div class="space-y-1">
        <p class="text-sm text-gray-500">Created</p>
        <p class="font-medium">{timestampDate(campaign.createdAt).toLocaleString()}</p>
      </div>
    {/if}

    <div class="space-y-2">
      <Heading tag="h3" class="text-lg">Members</Heading>
      <Listgroup aria-label="Members">
        {#each campaign.members as member (member.id)}
          <ListgroupItem>{member.name}</ListgroupItem>
        {/each}
      </Listgroup>
    </div>
  </div>
{:else}
  <div class="px-5">
    <ErrorMessage title="Campaign" error={errorObj} />
  </div>
{/if}

<Footer />
