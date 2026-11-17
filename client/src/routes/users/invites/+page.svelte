<script lang="ts">
  import Page from '$lib/components/page.svelte';
  import ActionButton from '$lib/components/buttons/action-button.svelte';
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { inviteHandler } from '$lib/handlers/invites';
  import { userHandler } from '$lib/handlers/user';
  import { timestampDate } from '@bufbuild/protobuf/wkt';
  import {
    Button,
    Input,
    Label,
    Modal,
    Select,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell
  } from 'flowbite-svelte';
  import QRCode from 'qrcode';
  import { onMount } from 'svelte';
  import type { Invite } from '../../../proto/users/v1/invites_pb';

  const HOUR = 60 * 60;
  const DURATIONS = [
    { value: HOUR, name: '1 hour' },
    { value: 24 * HOUR, name: '1 day' },
    { value: 7 * 24 * HOUR, name: '7 days' },
    { value: 30 * 24 * HOUR, name: '30 days' }
  ];

  let invites = $state<Invite[]>([]);
  let validFor = $state(7 * 24 * HOUR);
  let errorObj = $state<MechanusError>(null);

  // The invite shown with its link and QR code
  let shown = $state<Invite | undefined>(undefined);
  let shownOpen = $state(false);
  let shownLink = $derived(shown ? inviteHandler.link(shown.id) : '');
  let qrCode = $state('');
  let copied = $state(false);

  function toError(err: unknown): MechanusError {
    return err instanceof Error ? err : new Error('An unexpected error occurred.');
  }

  function expiresAt(invite: Invite): string {
    return invite.expiresAt ? timestampDate(invite.expiresAt).toLocaleString() : '';
  }

  async function load() {
    try {
      invites = await inviteHandler.list();
    } catch (err) {
      errorObj = toError(err);
    }
  }

  async function show(invite: Invite) {
    shown = invite;
    copied = false;
    qrCode = await QRCode.toDataURL(inviteHandler.link(invite.id), { margin: 1, width: 256 });
    shownOpen = true;
  }

  async function create(event: Event) {
    event.preventDefault();
    errorObj = null;

    try {
      const invite = await inviteHandler.create(validFor);
      await load();
      if (invite) await show(invite);
    } catch (err) {
      errorObj = toError(err);
    }
  }

  async function revoke(invite: Invite) {
    errorObj = null;

    try {
      await inviteHandler.revoke(invite.id);
      await load();
    } catch (err) {
      errorObj = toError(err);
    }
  }

  async function copy() {
    try {
      await navigator.clipboard.writeText(shownLink);
      copied = true;
    } catch (err) {
      // The clipboard API needs a secure context, which a LAN server over plain HTTP isn't
      console.error('could not copy the invite link', err);
    }
  }

  onMount(() => {
    if (!userHandler.current.loggedin) {
      return goto('/users/login');
    }
    if (!userHandler.isAdmin) {
      return goto('/users/profile');
    }

    return load();
  });
</script>

<svelte:head>
  <title>Invites</title>
</svelte:head>

<NavBar />

<Page>
  <div class="space-y-4">
    <h1 class="text-ink text-2xl font-bold">Invites</h1>
    <p class="text-muted">
      An invite is a single use link a new player opens to pick their own username and password.
    </p>

    <form class="panel flex flex-wrap items-end gap-3" onsubmit={create}>
      <Label class="space-y-2">
        <span class="text-ink">Valid for</span>
        <Select items={DURATIONS} bind:value={validFor} />
      </Label>
      <ActionButton intent="create" type="submit">Create invite</ActionButton>
    </form>

    <ErrorMessage error={errorObj} />

    <Table class="border-line overflow-hidden rounded-lg border">
      <TableHead>
        <TableHeadCell>Created</TableHeadCell>
        <TableHeadCell>Expires</TableHeadCell>
        <TableHeadCell><span class="sr-only">Actions</span></TableHeadCell>
      </TableHead>
      <TableBody>
        {#each invites as invite (invite.id)}
          <TableBodyRow data-testid={`invite-${invite.id}`}>
            <TableBodyCell
              >{invite.createdAt
                ? timestampDate(invite.createdAt).toLocaleString()
                : ''}</TableBodyCell
            >
            <TableBodyCell>{expiresAt(invite)}</TableBodyCell>
            <TableBodyCell class="space-x-1 text-right">
              <Button size="xs" color="alternative" onclick={() => show(invite)}>Show</Button>
              <ActionButton intent="delete" size="xs" onclick={() => revoke(invite)}
                >Revoke</ActionButton
              >
            </TableBodyCell>
          </TableBodyRow>
        {:else}
          <TableBodyRow>
            <TableBodyCell colspan={3}>No open invites</TableBodyCell>
          </TableBodyRow>
        {/each}
      </TableBody>
    </Table>
  </div>
</Page>

<Modal title="Invite" bind:open={shownOpen} size="sm">
  <div class="space-y-4">
    {#if qrCode}
      <img src={qrCode} alt="QR code of the invite link" class="mx-auto" width="256" height="256" />
    {/if}
    <Label class="space-y-2">
      <span>Invite link</span>
      <Input type="text" readonly value={shownLink} data-testid="invite-link" />
    </Label>
    {#if shown}
      <p class="text-muted text-sm">Single use, expires {expiresAt(shown)}.</p>
    {/if}
    <Button class="w-full" onclick={copy}>{copied ? 'Copied' : 'Copy link'}</Button>
  </div>
</Modal>

<Footer />
