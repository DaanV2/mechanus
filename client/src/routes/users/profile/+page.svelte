<script lang="ts">
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import { Button, Input, Label } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import type { User } from '../../../proto/users/v1/users_pb';

  let user = $state<User | undefined>(undefined);

  let currentPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');
  let passwordError = $state<MechanusError>(null);
  let passwordChanged = $state(false);
  let passwordsMatch = $derived(newPassword === confirmPassword);
  let canChangePassword = $derived(
    currentPassword !== '' && newPassword.trim() !== '' && passwordsMatch
  );

  onMount(async () => {
    const current = await userHandler.restore();
    if (!current.loggedin) {
      // redirect to login
      return goto('/users/login');
    }

    try {
      const data = await userHandler.serverData();
      user = data.user;
    } catch (err) {
      // A revoked session fails to refresh, which forgets its tokens
      if (!userHandler.current.loggedin) return goto('/users/login');

      throw err;
    }
  });

  async function changePassword(event: Event) {
    event.preventDefault();
    passwordError = null;
    passwordChanged = false;
    if (!canChangePassword) return;

    try {
      await userHandler.changePassword(currentPassword, newPassword);
      passwordChanged = true;
      currentPassword = newPassword = confirmPassword = '';
    } catch (err) {
      passwordError = err instanceof Error ? err : new Error('An unexpected error occurred.');
    }
  }
</script>

<svelte:head>
  {#if user}
    <title>User - {user.name}</title>
  {:else}
    <title>User</title>
  {/if}
</svelte:head>

<NavBar />

<div class="flex min-h-screen flex-col items-center justify-center gap-4">
  <div
    class="w-full max-w-md space-y-4 rounded-lg border border-gray-700 bg-gray-800 p-8 shadow-lg"
  >
    <h1 class="text-2xl font-bold text-white">Welcome{user ? `, ${user.name}` : ''}!</h1>
    {#if user}
      <div class="space-y-2">
        <p class="text-sm text-gray-400">Username</p>
        <p class="font-medium text-white">{user.name}</p>
      </div>
    {/if}
  </div>

  {#if user}
    <form
      class="w-full max-w-md space-y-4 rounded-lg border border-gray-700 bg-gray-800 p-8 shadow-lg"
      onsubmit={changePassword}
    >
      <h2 class="text-xl font-bold text-white">Change password</h2>
      <p class="text-sm text-gray-400">This signs you out on all your other devices.</p>
      <Label class="space-y-2">
        <span class="font-bold text-white">Current password</span>
        <Input type="password" required bind:value={currentPassword} />
      </Label>
      <Label class="space-y-2">
        <span class="font-bold text-white">New password</span>
        <Input type="password" required bind:value={newPassword} />
      </Label>
      <Label class="space-y-2">
        <span class="font-bold {passwordsMatch ? 'text-white' : 'text-red-800'}"
          >Confirm new password</span
        >
        <Input type="password" required bind:value={confirmPassword} />
      </Label>
      <Button type="submit" class="w-full" disabled={!canChangePassword}>Change password</Button>
      {#if passwordChanged}
        <p class="text-sm text-green-400">Password changed.</p>
      {/if}
      <ErrorMessage error={passwordError} title="Changing password failed" />
    </form>
  {/if}
</div>

<Footer />
