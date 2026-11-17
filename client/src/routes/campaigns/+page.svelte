<script lang="ts">
  import Page from '$lib/components/page.svelte';
  import ActionButton from '$lib/components/buttons/action-button.svelte';
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import { createCampaignClient } from '$lib/stores/clients';
  import { Input, Label } from 'flowbite-svelte';
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

<Page title={isAdmin ? 'All campaigns' : 'Your campaigns'}>
  <form class="panel mb-6 flex flex-wrap items-end gap-3" onsubmit={handleCreate}>
    <div class="min-w-48 grow space-y-2">
      <Label for="campaign-name" class="text-ink dark:text-ink font-bold">Campaign name</Label>
      <Input id="campaign-name" placeholder="The Fellowship" bind:value={name} />
    </div>
    <ActionButton intent="create" type="submit" disabled={!isFormValid}>Create</ActionButton>
  </form>

  <ErrorMessage title="Campaigns" error={errorObj} />

  {#if campaigns === undefined}
    <p class="text-muted p-4">Loading campaigns...</p>
  {:else if campaigns.length === 0}
    <p class="text-muted p-4">You are not in any campaigns yet.</p>
  {:else}
    <ul
      class="bg-surface border-line divide-line divide-y overflow-hidden rounded-lg border"
      aria-label="Campaigns"
    >
      {#each campaigns as campaign (campaign.id)}
        <li>
          <a
            href={`/campaigns/view/?id=${encodeURIComponent(campaign.id)}`}
            class="hover:bg-page group flex items-center justify-between gap-4 px-5 py-4 transition-colors"
          >
            <span class="text-ink group-hover:text-accent font-semibold">{campaign.name}</span>
            <span class="text-muted flex shrink-0 items-center gap-2 text-sm">
              {campaign.members.length} member{campaign.members.length === 1 ? '' : 's'}
              <span aria-hidden="true">&rsaquo;</span>
            </span>
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</Page>

<Footer />
