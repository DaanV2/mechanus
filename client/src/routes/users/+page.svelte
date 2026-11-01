<script lang="ts">
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import {
    Button,
    Checkbox,
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

  /** The roles an admin can hand out, from most to least permissions */
  const ROLES = ['admin', 'operator', 'user', 'viewer'];

  let users = $state<User[]>([]);
  let errorObj = $state<MechanusError>(null);
  let success = $state('');

  function toError(err: unknown): MechanusError {
    return err instanceof Error ? err : new Error('An unexpected error occurred.');
  }

  // Reset password dialog state
  let resetUser = $state<User | undefined>(undefined);
  let resetOpen = $state(false);
  let newPassword = $state('');
  let confirmPassword = $state('');
  let resetError = $state<MechanusError>(null);
  let isResetValid = $derived(newPassword.trim() !== '' && newPassword === confirmPassword);

  // Create account dialog state
  let createOpen = $state(false);
  let createUsername = $state('');
  let createPassword = $state('');
  let createRoles = $state<string[]>(['user']);
  let createError = $state<MechanusError>(null);
  let isCreateValid = $derived(
    createUsername.trim() !== '' && createPassword.trim() !== '' && createRoles.length > 0
  );

  // Edit roles dialog state
  let rolesUser = $state<User | undefined>(undefined);
  let rolesOpen = $state(false);
  let editRoles = $state<string[]>([]);
  let rolesError = $state<MechanusError>(null);
  let isSelf = $derived(rolesUser?.id === userHandler.current.data?.user.id);
  // An admin can't remove their own admin role, the server refuses it too
  let isRolesValid = $derived(editRoles.length > 0 && (!isSelf || editRoles.includes('admin')));

  async function load() {
    try {
      users = await userHandler.list();
    } catch (err) {
      errorObj = toError(err);
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
      success = `Password of ${resetUser.name} has been reset`;
      resetOpen = false;
    } catch (err) {
      resetError = toError(err);
    }
  }

  function openCreate() {
    createUsername = '';
    createPassword = '';
    createRoles = ['user'];
    createError = null;
    createOpen = true;
  }

  async function handleCreate(event: Event) {
    event.preventDefault();
    if (!isCreateValid) return;
    createError = null;

    try {
      await userHandler.createAccount(createUsername.trim(), createPassword, createRoles);
      success = `Account ${createUsername.trim()} has been created`;
      createOpen = false;
      await load();
    } catch (err) {
      createError = toError(err);
    }
  }

  function openRoles(user: User) {
    rolesUser = user;
    editRoles = [...user.roles];
    rolesError = null;
    rolesOpen = true;
  }

  async function handleRoles(event: Event) {
    event.preventDefault();
    if (!rolesUser || !isRolesValid) return;
    rolesError = null;

    try {
      await userHandler.setRoles(rolesUser.id, editRoles);
      success = `Roles of ${rolesUser.name} have been updated, they apply once their session refreshes`;
      rolesOpen = false;
      await load();
    } catch (err) {
      rolesError = toError(err);
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
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h1 class="text-2xl font-bold text-white">Users</h1>
      <div class="flex gap-2">
        <Button color="alternative" href="/users/invites">Invites</Button>
        <Button onclick={openCreate}>Create account</Button>
      </div>
    </div>

    {#if success}
      <p role="status" class="rounded-lg bg-green-800 p-4 text-white">{success}</p>
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
            <TableBodyCell class="space-x-1 text-right">
              <Button
                size="xs"
                color="alternative"
                aria-label={`Edit roles of ${user.name}`}
                onclick={() => openRoles(user)}>Edit roles</Button
              >
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

<Modal title="Create account" bind:open={createOpen} size="sm">
  <form class="space-y-4" onsubmit={handleCreate}>
    <Label class="space-y-2">
      <span>Username</span>
      <Input type="text" required bind:value={createUsername} />
    </Label>
    <Label class="space-y-2">
      <span>Password</span>
      <Input type="password" required bind:value={createPassword} />
    </Label>
    <fieldset class="space-y-2">
      <legend class="text-sm font-medium">Roles</legend>
      {#each ROLES as role (role)}
        <Checkbox bind:group={createRoles} value={role}>{role}</Checkbox>
      {/each}
    </fieldset>
    <Button type="submit" class="w-full" disabled={!isCreateValid}>Create</Button>
    <ErrorMessage error={createError} />
  </form>
</Modal>

<Modal title={`Roles of ${rolesUser?.name ?? ''}`} bind:open={rolesOpen} size="sm">
  <form class="space-y-4" onsubmit={handleRoles}>
    <fieldset class="space-y-2">
      <legend class="sr-only">Roles</legend>
      {#each ROLES as role (role)}
        <Checkbox bind:group={editRoles} value={role} disabled={isSelf && role === 'admin'}
          >{role}</Checkbox
        >
      {/each}
    </fieldset>
    {#if isSelf}
      <p class="text-sm text-gray-500">You can't remove your own admin role.</p>
    {/if}
    <Button type="submit" class="w-full" disabled={!isRolesValid}>Save</Button>
    <ErrorMessage error={rolesError} />
  </form>
</Modal>

<Footer />
