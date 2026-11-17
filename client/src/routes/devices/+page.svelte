<script lang="ts">
  import Page from '$lib/components/page.svelte';
  import ActionButton from '$lib/components/buttons/action-button.svelte';
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { deviceHandler } from '$lib/handlers/devices';
  import { userHandler } from '$lib/handlers/user';
  import { timestampDate } from '@bufbuild/protobuf/wkt';
  import {
    Badge,
    Input,
    Label,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell
  } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import type { Device } from '../../proto/devices/v1/devices_pb';

  /** How often the list is updated, to show devices connecting and leaving */
  const POLL_INTERVAL_MS = 5000;

  let devices = $state<Device[]>([]);
  let code = $state('');
  let name = $state('');
  let errorObj = $state<MechanusError>(null);

  function toError(err: unknown): MechanusError {
    return err instanceof Error ? err : new Error('An unexpected error occurred.');
  }

  function lastSeen(device: Device): string {
    return device.lastSeenAt ? timestampDate(device.lastSeenAt).toLocaleString() : '';
  }

  async function load() {
    try {
      devices = await deviceHandler.list();
    } catch (err) {
      errorObj = toError(err);
    }
  }

  async function confirm(event: Event) {
    event.preventDefault();
    errorObj = null;

    try {
      await deviceHandler.confirm(code.trim(), name);
      code = '';
      name = '';
      await load();
    } catch (err) {
      errorObj = toError(err);
    }
  }

  async function revoke(device: Device) {
    errorObj = null;

    try {
      await deviceHandler.revoke(device.id);
      await load();
    } catch (err) {
      errorObj = toError(err);
    }
  }

  async function bind(device: Device, event: Event) {
    errorObj = null;

    try {
      await deviceHandler.bind(device.id, (event.target as HTMLInputElement).value);
      await load();
    } catch (err) {
      errorObj = toError(err);
    }
  }

  onMount(() => {
    if (!userHandler.current.loggedin) {
      goto('/users/login');

      return;
    }
    if (!userHandler.isAdmin) {
      goto('/users/profile');

      return;
    }

    load();
    const poll = setInterval(load, POLL_INTERVAL_MS);

    return () => clearInterval(poll);
  });
</script>

<svelte:head>
  <title>Devices</title>
</svelte:head>

<NavBar />

<Page width="wide">
  <div class="space-y-4">
    <h1 class="text-ink text-2xl font-bold">Devices</h1>
    <p class="text-muted">
      To pair a screen, open <code class="bg-surface text-ink rounded px-1">/devices/pair</code> on it
      and enter the code it shows.
    </p>

    <form class="panel flex flex-wrap items-end gap-3" onsubmit={confirm}>
      <Label class="space-y-2">
        <span class="text-ink">Pairing code</span>
        <Input
          type="text"
          inputmode="numeric"
          maxlength={4}
          pattern={'[0-9]{4}'}
          required
          autocomplete="off"
          bind:value={code}
        />
      </Label>
      <Label class="space-y-2">
        <span class="text-ink">Device name</span>
        <Input type="text" maxlength={64} required placeholder="Table TV" bind:value={name} />
      </Label>
      <ActionButton intent="create" type="submit">Pair device</ActionButton>
    </form>

    <ErrorMessage error={errorObj} title="Device error" />

    <Table class="border-line overflow-hidden rounded-lg border">
      <TableHead>
        <TableHeadCell>Name</TableHeadCell>
        <TableHeadCell>Hardware</TableHeadCell>
        <TableHeadCell>Target</TableHeadCell>
        <TableHeadCell>Status</TableHeadCell>
        <TableHeadCell>Last seen</TableHeadCell>
        <TableHeadCell><span class="sr-only">Actions</span></TableHeadCell>
      </TableHead>
      <TableBody>
        {#each devices as device (device.id)}
          <TableBodyRow data-testid={`device-${device.id}`}>
            <TableBodyCell>{device.name || device.id}</TableBodyCell>
            <TableBodyCell>{device.hardware}</TableBodyCell>
            <TableBodyCell>
              {#if device.paired}
                <Input
                  type="text"
                  size="sm"
                  aria-label={`Target of ${device.name}`}
                  placeholder="table"
                  value={device.target}
                  onchange={(event: Event) => bind(device, event)}
                />
              {/if}
            </TableBodyCell>
            <TableBodyCell>
              <Badge color={device.connected ? 'green' : 'gray'}>
                {device.connected ? 'Connected' : 'Offline'}
              </Badge>
              {#if !device.paired}
                <Badge color="yellow">Not paired</Badge>
              {/if}
            </TableBodyCell>
            <TableBodyCell>{lastSeen(device)}</TableBodyCell>
            <TableBodyCell class="text-right">
              {#if device.paired}
                <ActionButton intent="delete" size="xs" onclick={() => revoke(device)}
                  >Revoke</ActionButton
                >
              {/if}
            </TableBodyCell>
          </TableBodyRow>
        {:else}
          <TableBodyRow>
            <TableBodyCell colspan={6}>No devices yet</TableBodyCell>
          </TableBodyRow>
        {/each}
      </TableBody>
    </Table>
  </div>
</Page>

<Footer />
