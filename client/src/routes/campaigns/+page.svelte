<script lang="ts">
  import { goto } from '$app/navigation';
  import SimpleCard from '$lib/components/cards/simple.svelte';
  import Fluid from '$lib/components/collections/fluid.svelte';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import { createCampaignClient } from '$lib/stores/clients';
  import { Button, Heading, Hr, Input, Label } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import type { Campaign } from '../../proto/campaigns/v1/campaign_pb';

  let campaigns = $state<Campaign[] | undefined>(undefined);
  let isAdmin = $state(false);
  let name = $state('');
  let errorObj = $state<MechanusError>(null);
  let isFormValid = $derived(name.trim() !== '');

  async function load() {
    const resp = await createCampaignClient().list({});
    campaigns = resp.campaigns;
  }

  async function handleCreate(event: Event) {
    event.preventDefault();
    errorObj = null;
    if (!isFormValid) return;

    try {
      const resp = await createCampaignClient().create({ name: name.trim() });
      name = '';
      await goto(`/campaigns/view/?id=${encodeURIComponent(resp.campaign?.id ?? '')}`);
    } catch (err) {
      errorObj = err instanceof Error ? err : new Error('An unexpected error occurred.');
    }
  }

  onMount(async () => {
    const current = userHandler.current;
    if (!current.loggedin) {
      return goto('/users/login');
    }
    isAdmin = current.data.user.roles?.includes('admin') ?? false;

    try {
      await load();
    } catch (err) {
      errorObj = err instanceof Error ? err : new Error('An unexpected error occurred.');
    }
  });
</script>

<svelte:head>
  <title>Campaigns</title>
</svelte:head>

<NavBar />

<Heading tag="h2" class="px-5 py-3">{isAdmin ? 'All campaigns' : 'Your campaigns'}</Heading>
<Hr />

<div class="flex flex-col items-center">
  <form class="flex w-full max-w-md items-end gap-2 px-5" onsubmit={handleCreate}>
    <Label class="grow space-y-2">
      <Label for="campaign-name" class="font-bold">Campaign name</Label>
      <Input id="campaign-name" placeholder="The Fellowship" bind:value={name} />
    </Label>
    <Button type="submit" disabled={!isFormValid}>Create</Button>
  </form>

  <ErrorMessage title="Campaigns" error={errorObj} />

  {#if campaigns === undefined}
    <p class="p-4 text-gray-500">Loading campaigns...</p>
  {:else if campaigns.length === 0}
    <p class="p-4 text-gray-500">You are not in any campaigns yet.</p>
  {:else}
    <Fluid>
      {#each campaigns as campaign (campaign.id)}
        <SimpleCard
          href={`/campaigns/view/?id=${encodeURIComponent(campaign.id)}`}
          title={campaign.name}
          text={`${campaign.members.length} member${campaign.members.length === 1 ? '' : 's'}`}
        />
      {/each}
    </Fluid>
  {/if}
</div>

<Footer />
