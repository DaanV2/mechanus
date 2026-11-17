<script lang="ts">
  import Page from '$lib/components/page.svelte';
  import ActionButton from '$lib/components/buttons/action-button.svelte';
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import { Input, Label } from 'flowbite-svelte';
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

<Page width="narrow" class="flex flex-col justify-center gap-4">
  <div class="panel w-full max-w-md space-y-4">
    <h1 class="text-ink text-2xl font-bold">Welcome{user ? `, ${user.name}` : ''}!</h1>
    {#if user}
      <div class="space-y-2">
        <p class="text-muted text-sm">Username</p>
        <p class="text-ink font-medium">{user.name}</p>
      </div>
    {/if}
  </div>

  {#if user}
    <form class="panel w-full max-w-md space-y-4" onsubmit={changePassword}>
      <h2 class="text-ink text-xl font-bold">Change password</h2>
      <p class="text-muted text-sm">This signs you out on all your other devices.</p>
      <Label class="space-y-2">
        <span class="text-ink font-bold">Current password</span>
        <Input type="password" required bind:value={currentPassword} />
      </Label>
      <Label class="space-y-2">
        <span class="text-ink font-bold">New password</span>
        <Input type="password" required bind:value={newPassword} />
      </Label>
      <Label class="space-y-2">
        <span class="font-bold {passwordsMatch ? 'text-ink' : 'text-error-700 dark:text-error-400'}"
          >Confirm new password</span
        >
        <Input type="password" required bind:value={confirmPassword} />
      </Label>
      <ActionButton type="submit" class="w-full" disabled={!canChangePassword}
        >Change password</ActionButton
      >
      {#if passwordChanged}
        <p class="text-ok-700 dark:text-ok-400 text-sm">Password changed.</p>
      {/if}
      <ErrorMessage error={passwordError} title="Changing password failed" />
    </form>
  {/if}
</Page>

<Footer />
