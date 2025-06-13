<script lang="ts">
  import { redirect } from '@sveltejs/kit';
  import { createClient } from '../../../lib/api/client';
  import { createLoginClient } from '../../../lib/api/users_v1';
  import { ConnectError } from '@connectrpc/connect';
  import ErrorMessage from '$lib/components/error-message.svelte';

  let username = '';
  let password = '';
  let errorObj: ConnectError | Error | null = null;

  $: isFormValid = username.trim() !== '' && password.trim() !== '';

  async function handleSubmit(event: Event) {
    event.preventDefault();
    errorObj = null;
    if (!isFormValid) return;

    const transport = createClient();
    const loginClient = createLoginClient(transport);

    try {
      const login = await loginClient.login({ username, password });
      redirect(302, '/users/profile');
    } catch (err) {
      if (err instanceof ConnectError) {
        errorObj = err;
      } else if (err instanceof Error) {
        errorObj = err;
      } else {
        errorObj = new Error('An unexpected error occurred.');
      }
    }
  }
</script>

<svelte:head>
  <title>LargestContentfulPaint</title>
</svelte:head>

<div class="centered-container">
  <form class="box-container" on:submit={handleSubmit}>
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
