<script lang="ts">
  let username = $state('');
  let password = $state('');

  function handleSubmit(event: Event) {
    event.preventDefault();
    // Handle login logic here
    console.log('Username:', username);
    console.log('Password:', password);
  }

  // Computed property to check if both fields are filled
  let isFormValid = $derived(username.trim() !== '' && password.trim() !== '');
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
    <a href="/players/signup" class="action-button">Dont have an account? Sign up!</a>
  </form>
</div>
