<script lang="ts">
  import { goto } from '$app/navigation';
  import { Code, ConnectError } from '@connectrpc/connect';
  import ErrorMessage from '../../../lib/components/error-message.svelte';
  import type { MechanusError } from '../../../lib/components/errors.svelte';
  import { KEY_ACCESS_TOKEN, setCookie } from '../../../lib/cookies';
  import { createLoginClient, createUserClient } from '../../../lib/stores/clients';

  let username = $state('');
  let password = $state('');
  let confirm_password = $state('');
  let errorObj = $state<MechanusError>(null);

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

    await createUserClient().create({ username, password });
    const login = await createLoginClient().login({ username, password });
    setCookie(KEY_ACCESS_TOKEN, `${login.type} ${login.token}`);

    goto('/users/profile');
  }

  // Computed property to check if both fields are filled
  let correct_password = $derived(password === confirm_password);
  let isFormValid = $derived(username.trim() !== '' && password.trim() !== '' && correct_password);
  let password_wrong = $derived(password.trim() !== '' && !correct_password);
</script>

<svelte:head>
  <title>Signup</title>
</svelte:head>

<div class="centered-container">
  <form class="box-container" onsubmit={handleSubmit}>
    <input type="text" class="login-input" placeholder="Username" bind:value={username} required />
    <input
      type="password"
      class="login-input p-2 border rounded resizable-box"
      placeholder="Password"
      bind:value={password}
      required
    />
    <input
      type="password"
      class="login-input p-2 border rounded resizable-box"
      placeholder="Confirm Password"
      bind:value={confirm_password}
      required
    />
    {#if password_wrong && confirm_password.length > 0}
      <p class="text-red-500 resizable-box">The password doesn't match the confirmed password</p>
    {/if}

    <button type="submit" class="action-button" disabled={!isFormValid}> Signup </button>
    <a href="/users/login" class="action-button">Already have an account?</a>
    <ErrorMessage error={errorObj} />
  </form>
</div>

<style>
  .resizable-box {
    transition:
      width 0.5s ease,
      height 0.5s ease;
  }
</style>
