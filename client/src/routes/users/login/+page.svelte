<script lang="ts">
  import { goto } from '$app/navigation';
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
  let errorObj = $state<MechanusError>(null);
  let showPassword = $state(false);
  let isFormValid = $derived(username.trim() !== '' && password.trim() !== '');

  async function handleSubmit(event: Event) {
    event.preventDefault();
    errorObj = null;
    if (!isFormValid) return;

    return login().catch((err) => {
      if (err instanceof ConnectError) {
        errorObj = err;

        if (err.code == Code.Unauthenticated) {
          errorObj = 'wrong password / username';
        }
      } else if (err instanceof Error) {
        errorObj = err;
      } else {
        errorObj = new Error('An unexpected error occurred.');
      }
    });
  }

  async function login() {
    if (!isFormValid) return;

    await userHandler.login(username, password);
    await sleep(100);
    goto('/users/profile');
  }

  onMount(() => {
    if (userHandler.current.loggedin) {
      console.error('already logged in');
      goto('/users/profile');
    }
  });
</script>

<svelte:head>
  <title>Login</title>
</svelte:head>

<NavBar />

<div class="flex min-h-screen items-center justify-center">
  <form class="space-y-6" onsubmit={handleSubmit}>
    <h3 class="p-0 text-xl font-medium text-white dark:text-white">Login</h3>
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
    <Button type="submit" class="w-full1" disabled={!isFormValid}>Login</Button>
    <p class="text-sm font-light text-white dark:text-white">
      Don't have an account yet? <a
        href="/users/signup"
        class="text-primary-600 dark:text-primary-500 font-medium hover:underline">Sign up</a
      >
    </p>

    <ErrorMessage error={errorObj} />
  </form>
</div>

<Footer />
