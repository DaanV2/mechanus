<script lang="ts">
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import {
    Button,
    Input,
    Label,
    Modal,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell
  } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import type { User } from '../../proto/users/v1/users_pb';

  let users = $state<User[]>([]);
  let errorObj = $state<MechanusError>(null);

  // Reset password dialog state
  let resetUser = $state<User | undefined>(undefined);
  let resetOpen = $state(false);
  let newPassword = $state('');
  let confirmPassword = $state('');
  let resetError = $state<MechanusError>(null);
  let resetSuccess = $state('');
  let isResetValid = $derived(newPassword.trim() !== '' && newPassword === confirmPassword);

  async function load() {
    try {
      users = await userHandler.list();
    } catch (err) {
      errorObj = err instanceof Error ? err : new Error('An unexpected error occurred.');
    }
  }

  function openReset(user: User) {
    resetUser = user;
    newPassword = '';
    confirmPassword = '';
    resetError = null;
    resetOpen = true;
  }

  async function handleReset(event: Event) {
    event.preventDefault();
    if (!resetUser || !isResetValid) return;
    resetError = null;

    try {
      await userHandler.resetPassword(resetUser.id, newPassword);
      resetSuccess = `Password of ${resetUser.name} has been reset`;
      resetOpen = false;
    } catch (err) {
      resetError = err instanceof Error ? err : new Error('An unexpected error occurred.');
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
  <title>Users</title>
</svelte:head>

<NavBar />

<div class="flex min-h-screen flex-col items-center p-8">
  <div class="w-full max-w-3xl space-y-4">
    <h1 class="text-2xl font-bold text-white">Users</h1>

    {#if resetSuccess}
      <p role="status" class="rounded-lg bg-green-800 p-4 text-white">{resetSuccess}</p>
    {/if}
    <ErrorMessage error={errorObj} />

    <Table>
      <TableHead>
        <TableHeadCell>Username</TableHeadCell>
        <TableHeadCell>Roles</TableHeadCell>
        <TableHeadCell><span class="sr-only">Actions</span></TableHeadCell>
      </TableHead>
      <TableBody>
        {#each users as user (user.id)}
          <TableBodyRow>
            <TableBodyCell>{user.name}</TableBodyCell>
            <TableBodyCell>{user.roles.join(', ')}</TableBodyCell>
            <TableBodyCell>
              <Button
                size="xs"
                aria-label={`Reset password of ${user.name}`}
                onclick={() => openReset(user)}>Reset password</Button
              >
            </TableBodyCell>
          </TableBodyRow>
        {/each}
      </TableBody>
    </Table>
  </div>
</div>

<Modal title={`Reset password of ${resetUser?.name ?? ''}`} bind:open={resetOpen} size="sm">
  <form class="space-y-4" onsubmit={handleReset}>
    <Label class="space-y-2">
      <span>New password</span>
      <Input type="password" required bind:value={newPassword} />
    </Label>
    <Label class="space-y-2">
      <span>Confirm new password</span>
      <Input type="password" required bind:value={confirmPassword} />
    </Label>
    <Button type="submit" class="w-full" disabled={!isResetValid}>Reset</Button>
    <ErrorMessage error={resetError} />
  </form>
</Modal>

<Footer />
