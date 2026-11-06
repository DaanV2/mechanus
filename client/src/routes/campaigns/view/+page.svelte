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
  import {
    Button,
    Heading,
    Hr,
    Input,
    Label,
    Listgroup,
    ListgroupItem,
    Modal,
    Textarea
  } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import type { Campaign, CampaignMember } from '../../../proto/campaigns/v1/campaign_pb';

  let campaign = $state<Campaign | undefined>(undefined);
  let errorObj = $state<MechanusError>(null);
  let currentUserId = $state('');
  let isAdmin = $state(false);

  // Edit form state
  let editName = $state('');
  let editDescription = $state('');
  let editError = $state<MechanusError>(null);
  let editSuccess = $state('');
  let isEditValid = $derived(editName.trim() !== '');

  // Players state
  let addUsername = $state('');
  let playersError = $state<MechanusError>(null);
  let isAddValid = $derived(addUsername.trim() !== '');

  // Delete dialog state
  let deleteOpen = $state(false);
  let deleteError = $state<MechanusError>(null);

  function toError(err: unknown): MechanusError {
    if (err instanceof ConnectError) {
      return err.rawMessage;
    }
    return err instanceof Error ? err : new Error('An unexpected error occurred.');
  }

  function setCampaign(value: Campaign | undefined) {
    campaign = value;
    editName = value?.name ?? '';
    editDescription = value?.description ?? '';
  }

  async function reload() {
    if (!campaign) return;
    const resp = await createCampaignClient().get({ id: campaign.id });
    setCampaign(resp.campaign);
  }

  async function handleSave(event: Event) {
    event.preventDefault();
    editError = null;
    editSuccess = '';
    if (!campaign || !isEditValid) return;

    try {
      const resp = await createCampaignClient().update({
        id: campaign.id,
        name: editName.trim(),
        description: editDescription.trim()
      });
      setCampaign(resp.campaign);
      editSuccess = 'Campaign saved';
    } catch (err) {
      editError = toError(err);
    }
  }

  async function handleAddPlayer(event: Event) {
    event.preventDefault();
    playersError = null;
    if (!campaign || !isAddValid) return;

    try {
      await createCampaignClient().addPlayer({
        campaignId: campaign.id,
        username: addUsername.trim()
      });
      addUsername = '';
      await reload();
    } catch (err) {
      if (err instanceof ConnectError && err.code === Code.NotFound) {
        playersError = `No user named "${addUsername.trim()}"`;
      } else {
        playersError = toError(err);
      }
    }
  }

  async function handleRemovePlayer(member: CampaignMember) {
    playersError = null;
    if (!campaign) return;

    try {
      await createCampaignClient().removePlayer({ campaignId: campaign.id, userId: member.id });
      if (member.id === currentUserId) {
        // Left the campaign, it can't be seen anymore
        return goto('/campaigns/');
      }
      await reload();
    } catch (err) {
      playersError = toError(err);
    }
  }

  async function handleDelete() {
    deleteError = null;
    if (!campaign) return;

    try {
      await createCampaignClient().delete({ id: campaign.id });
      deleteOpen = false;
      await goto('/campaigns/');
    } catch (err) {
      deleteError = toError(err);
    }
  }

  /** Whenever the current user may remove the given member: managers remove players, players can leave, nobody removes the owner */
  function canRemove(member: CampaignMember): boolean {
    if (!campaign || member.id === campaign.ownerId) return false;

    return campaign.canManage || member.id === currentUserId;
  }

  onMount(async () => {
    const current = userHandler.current;
    if (!current.loggedin) {
      return goto('/users/login');
    }
    currentUserId = current.data.user.id;
    isAdmin = userHandler.isAdmin;

    // The page is prerendered, so the id is read client side from the query string
    const id = new URLSearchParams(window.location.search).get('id') ?? '';
    if (id === '') {
      errorObj = 'No campaign selected';
      return;
    }

    try {
      const resp = await createCampaignClient().get({ id });
      setCampaign(resp.campaign);
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

  <div class="max-w-md space-y-6 px-5">
    {#if campaign.description}
      <p class="whitespace-pre-line" data-testid="campaign-description">{campaign.description}</p>
    {/if}

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
          <ListgroupItem>
            <div class="flex w-full items-center justify-between gap-2">
              <span>
                <span>{member.name}</span>
                {#if member.id === campaign.ownerId}
                  <span class="text-sm text-gray-500">(owner)</span>
                {/if}
              </span>
              {#if canRemove(member)}
                <Button
                  size="xs"
                  color="alternative"
                  aria-label={member.id === currentUserId
                    ? 'Leave campaign'
                    : `Remove ${member.name}`}
                  onclick={() => handleRemovePlayer(member)}
                >
                  {member.id === currentUserId ? 'Leave' : 'Remove'}
                </Button>
              {/if}
            </div>
          </ListgroupItem>
        {/each}
      </Listgroup>

      {#if campaign.canManage}
        <form class="flex items-end gap-2" onsubmit={handleAddPlayer}>
          <Label class="grow space-y-2">
            <span>Username</span>
            <Input placeholder="Username of an existing user" bind:value={addUsername} />
          </Label>
          <Button type="submit" disabled={!isAddValid}>Add player</Button>
        </form>
        <p class="text-sm text-gray-500">
          {#if isAdmin}
            New to Mechanus? <a href="/users/invites" class="underline">Create an invite link</a>,
            then add them here once they have signed up.
          {:else}
            New to Mechanus? Ask an admin for an invite link, then add them here once they have
            signed up.
          {/if}
        </p>
      {/if}
      <ErrorMessage title="Players" error={playersError} />
    </div>

    {#if campaign.canManage}
      <Hr />
      <form class="space-y-4" onsubmit={handleSave}>
        <Heading tag="h3" class="text-lg">Edit campaign</Heading>
        <Label class="space-y-2">
          <span>Name</span>
          <Input required bind:value={editName} />
        </Label>
        <Label class="space-y-2">
          <span>Description</span>
          <Textarea class="w-full" rows={4} bind:value={editDescription} />
        </Label>
        <Button type="submit" disabled={!isEditValid}>Save</Button>
        {#if editSuccess}
          <p class="text-sm text-green-600" role="status">{editSuccess}</p>
        {/if}
        <ErrorMessage title="Edit campaign" error={editError} />
      </form>

      <Hr />
      <div class="space-y-2 pb-6">
        <Heading tag="h3" class="text-lg">Danger zone</Heading>
        <Button color="red" onclick={() => ((deleteError = null), (deleteOpen = true))}>
          Delete campaign
        </Button>
      </div>
    {/if}
  </div>

  <Modal title="Delete campaign" bind:open={deleteOpen} size="sm">
    <p>
      Delete <strong>{campaign.name}</strong>? Its players lose access to it. This can't be undone
      from the app.
    </p>
    <div class="flex gap-2">
      <Button color="red" onclick={handleDelete}>Delete</Button>
      <Button color="alternative" onclick={() => (deleteOpen = false)}>Cancel</Button>
    </div>
    <ErrorMessage error={deleteError} />
  </Modal>
{:else}
  <div class="px-5">
    <ErrorMessage title="Campaign" error={errorObj} />
  </div>
{/if}

<Footer />
