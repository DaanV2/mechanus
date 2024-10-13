<script lang="ts">
  let username = '';
  let password = '';

  function handleSubmit(event: Event) {
    event.preventDefault();
    // Handle login logic here
    console.log('Username:', username);
    console.log('Password:', password);
  }
</script>

<form class="login-form" on:submit={handleSubmit}>
  <input type="text" class="login-input" placeholder="Username" bind:value={username} required />
  <input
    type="password"
    class="login-input"
    placeholder="Password"
    bind:value={password}
    required
  />
  <button type="submit" class="login-button">Login</button>
</form>

<style>
  .login-form {
    display: flex;
    flex-direction: column;
    width: 300px;
    margin: 0 auto;
    padding: 20px;
    border: 1px solid #ccc;
    border-radius: 8px;
    background-color: #f9f9f9;
  }

  .login-input {
    margin-bottom: 10px;
    padding: 8px;
    font-size: 1rem;
    border: 1px solid #ccc;
    border-radius: 4px;
  }

  .login-button {
    padding: 10px;
    font-size: 1rem;
    cursor: pointer;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    transition: background-color 0.3s;
  }

  .login-button:hover {
    background-color: #0056b3;
  }
</style>
