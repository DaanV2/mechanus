<script lang="ts">
  let username = '';
  let password = '';

  function handleSubmit(event: Event) {
    event.preventDefault();
    // Handle login logic here
    console.log('Username:', username);
    console.log('Password:', password);
  }

  // Computed property to check if both fields are filled
  $: isFormValid = username.trim() !== '' && password.trim() !== '';
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
  </form>
</div>
