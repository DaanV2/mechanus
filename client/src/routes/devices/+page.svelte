<script lang="ts">
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
    Button,
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

<div class="flex min-h-screen flex-col items-center p-8">
  <div class="w-full max-w-4xl space-y-4">
    <h1 class="text-2xl font-bold text-white">Devices</h1>
    <p class="text-gray-300">
      To pair a screen, open <code>/devices/pair</code> on it and enter the code it shows.
    </p>

    <form class="flex flex-wrap items-end gap-2" onsubmit={confirm}>
      <Label class="space-y-2">
        <span class="text-white">Pairing code</span>
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
        <span class="text-white">Device name</span>
        <Input type="text" maxlength={64} required placeholder="Table TV" bind:value={name} />
      </Label>
      <Button type="submit">Pair device</Button>
    </form>

    <ErrorMessage error={errorObj} title="Device error" />

    <Table>
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
                <Button size="xs" color="red" onclick={() => revoke(device)}>Revoke</Button>
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
</div>

<Footer />
