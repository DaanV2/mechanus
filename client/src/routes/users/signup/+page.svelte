<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { INVITE_PARAM, inviteHandler } from '$lib/handlers/invites';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { Code, ConnectError } from '@connectrpc/connect';
  import { Button, ButtonGroup, Input, InputAddon, Label } from 'flowbite-svelte';
  import { EyeOutline, EyeSlashOutline } from 'flowbite-svelte-icons';
  import { onMount } from 'svelte';
  import { userHandler } from '../../../lib/handlers/user';
  import { sleep } from '../../../lib/timings/sleep';

  let username = $state('');
  let password = $state('');
  let confirm_password = $state('');
  let errorObj = $state<MechanusError>(null);
  let showPassword = $state(false);
  let showConfirmPassword = $state(false);

  // Signing up with an invite link (`?invite=<id>`) instead of open sign-up
  let invite = $state('');
  let inviteExpires = $state<Date | undefined>(undefined);
  let inviteInvalid = $state(false);

  // Computed property to check if both fields are filled
  let correct_password = $derived(password === confirm_password);
  let isFormValid = $derived(
    username.trim() !== '' && password.trim() !== '' && correct_password && !inviteInvalid
  );
  let password_wrong = $derived(password.trim() !== '' && !correct_password);

  async function handleSubmit(event: Event) {
    event.preventDefault();
    errorObj = null;
    if (!isFormValid || password_wrong) return;

    return signup().catch((err) => {
      if (err instanceof ConnectError) {
        errorObj = err;

        if (err.code === Code.AlreadyExists) {
          errorObj = err.message;
        }
      } else if (err instanceof Error) {
        errorObj = err;
      } else {
        errorObj = new Error('An unexpected error occurred.');
      }
      return;
    });
  }

  async function signup() {
    if (password != confirm_password) {
      throw new Error('passwords not the same');
    }

    if (invite) {
      await inviteHandler.accept(invite, username, password);
    } else {
      await userHandler.create(username, password);
    }

    await sleep(100);
    goto('/users/profile');
  }

  onMount(() => {
    if (userHandler.current.loggedin) {
      console.error('already logged in for sign up');
      goto('/users/profile');

      return;
    }

    invite = page.url.searchParams.get(INVITE_PARAM) ?? '';
    if (invite) {
      inviteHandler
        .check(invite)
        .then((expires) => {
          inviteExpires = expires;
          inviteInvalid = expires === undefined;
        })
        .catch((err) => {
          console.error('could not check the invite', err);
          inviteInvalid = true;
        });
    }
  });
</script>

<svelte:head>
  <title>Signup</title>
</svelte:head>

<NavBar />

<div class="flex min-h-screen items-center justify-center">
  <form class="space-y-6" onsubmit={handleSubmit}>
    <h3 class="p-0 text-xl font-medium text-white dark:text-white">Signup</h3>
    {#if inviteInvalid}
      <p role="alert" class="rounded-lg bg-red-800 p-4 text-white">
        This invite is invalid, expired or already used. Ask your game master for a new one.
      </p>
    {:else if inviteExpires}
      <p class="text-sm text-white">
        You were invited, this invite expires {inviteExpires.toLocaleString()}.
      </p>
    {/if}
    <Label class="space-y-2">
      <Label for="username" class="font-bold text-white">Your username</Label>
      <Input
        id="username"
        type="username"
        name="username"
        placeholder="monotron"
        required
        bind:value={username}
      />
    </Label>
    <Label class="space-y-2">
      <Label for="password" class="font-bold text-white">Your password</Label>
      <ButtonGroup class="w-full">
        <Input
          id="password"
          type={showPassword ? 'text' : 'password'}
          placeholder="Your password here"
          required
          bind:value={password}
        />
        <InputAddon class="rounded-r-lg">
          <button onclick={() => (showPassword = !showPassword)}>
            {#if showPassword}
              <EyeOutline class="h-6 w-6" />
            {:else}
              <EyeSlashOutline class="h-6 w-6" />
            {/if}
          </button>
        </InputAddon>
      </ButtonGroup>
    </Label>
    <Label class="space-y-2">
      <Label
        for="confirm-password"
        class="font-bold {password_wrong ? 'text-red-800' : 'text-white'}">Confirm password</Label
      >
      <ButtonGroup class="w-full">
        <Input
          id="confirm-password"
          type={showConfirmPassword ? 'text' : 'password'}
          placeholder="Confirm your password here"
          color={password_wrong ? 'red' : confirm_password.length > 0 ? 'green' : 'default'}
          required
          bind:value={confirm_password}
        />
        <InputAddon class="rounded-r-lg">
          <button onclick={() => (showConfirmPassword = !showConfirmPassword)}>
            {#if showConfirmPassword}
              <EyeOutline class="h-6 w-6" />
            {:else}
              <EyeSlashOutline class="h-6 w-6" />
            {/if}
          </button>
        </InputAddon>
      </ButtonGroup>
    </Label>
    <Button type="submit" class="w-full" disabled={!isFormValid}>Signup</Button>
    <p class="text-sm font-light text-white dark:text-white">
      Already have an account? <a
        href="/users/login"
        class="text-primary-600 dark:text-primary-500 font-medium hover:underline">Login</a
      >
    </p>

    <ErrorMessage error={errorObj} />
  </form>
</div>

<Footer />
