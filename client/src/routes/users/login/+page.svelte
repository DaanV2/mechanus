<script lang="ts">
  import { goto } from '$app/navigation';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import { Code, ConnectError } from '@connectrpc/connect';
  import { createClient } from '../../../lib/api/client';
  import { createLoginClient } from '../../../lib/api/users_v1';
  import type { MechanusError } from '../../../lib/components/errors.svelte';
  import { KEY_ACCESS_TOKEN, setCookie } from '../../../lib/cookies';

  let username = $state('');
  let password = $state('');
  let errorObj = $state<MechanusError>(null);

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

    const transport = createClient();
    const loginClient = createLoginClient(transport);
    const login = await loginClient.login({ username, password });
    setCookie(KEY_ACCESS_TOKEN, `${login.type} ${login.token}`);

    goto('/users/profile');
  }
</script>

<svelte:head>
  <title>LargestContentfulPaint</title>
</svelte:head>

<div class="centered-container">
  <form class="box-container" onsubmit={handleSubmit}>
    <input type="text" class="login-input" placeholder="Username" bind:value={username} required />
    <input
      type="password"
      class="login-input"
      placeholder="Password"
      bind:value={password}
      required
    />
    <button type="submit" class="action-button" disabled={!isFormValid}> Login </button>
    <a href="/users/signup" class="action-button">Don't have an account? Sign up!</a>
    <ErrorMessage error={errorObj} />
  </form>
</div>
