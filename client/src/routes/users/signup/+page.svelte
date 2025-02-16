<script lang="ts">
  let username = $state('');
  let password = $state('');
  let confirm_password = $state('');

  function handleSubmit(event: Event) {
    event.preventDefault();
    // Handle login logic here
    console.log('Username:', username);
    console.log('Password:', password);
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
      class="login-input resizable-box"
      placeholder="Password"
      bind:value={password}
      required
    />
    <input
      type="password"
      class="p-2 border border-gray-300 rounded resizable-box"
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
