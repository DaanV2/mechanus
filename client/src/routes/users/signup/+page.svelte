<script lang="ts">
  import type { Client } from '@connectrpc/connect';
  import { onMount } from 'svelte';
  import { createLoginClient, createUserClient } from '../../../lib/stores/clients';
  import type { UserService } from '../../../proto/users/v1/users_connect';

  let username = $state('');
  let password = $state('');
  let confirm_password = $state('');

  let userClient: Client<typeof UserService>;

  onMount(() => {
    userClient = createUserClient();
  });

  async function handleSubmit(event: Event) {
    if (!userClient) {
      throw new Error("couldn't grab login rpcs");
    }
    if (password != confirm_password) {
      throw new Error('passwords not the same');
    }

    const response = await userClient.create({ username, password });
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
    {#if password_wrong}
      <p class="text-red-500 resizable-box">The password doesn't match the confirmed password</p>
    {/if}

    <button type="submit" class="action-button" disabled={!isFormValid}> Login </button>
  </form>
</div>

<style>
  .resizable-box {
    transition:
      width 0.5s ease,
      height 0.5s ease;
  }
</style>
