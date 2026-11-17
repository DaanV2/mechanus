<script lang="ts">
  import Page from '$lib/components/page.svelte';
  import ActionButton from '$lib/components/buttons/action-button.svelte';
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { Code, ConnectError } from '@connectrpc/connect';
  import { ButtonGroup, Input, InputAddon, Label } from 'flowbite-svelte';
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

<Page width="narrow" class="flex flex-col justify-center">
  <form class="panel space-y-6" onsubmit={handleSubmit}>
    <h3 class="text-ink p-0 text-xl font-medium">Login</h3>
    <Label class="space-y-2">
      <Label for="username" class="text-ink dark:text-ink font-bold">Your username</Label>
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
      <Label for="password" class="text-ink dark:text-ink font-bold">Your password</Label>
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
    <ActionButton type="submit" class="w-full" disabled={!isFormValid}>Login</ActionButton>
    <p class="text-muted text-sm">
      Don't have an account yet? <a href="/users/signup" class="link">Sign up</a>
    </p>

    <ErrorMessage error={errorObj} />
  </form>
</Page>

<Footer />
