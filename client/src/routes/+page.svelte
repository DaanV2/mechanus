<script lang="ts">
  import { onMount } from 'svelte';

  onMount(() => {
    console.log('cookies', document.cookie);
  });
</script>

<svelte:head>
  <title>Mechanus</title>
</svelte:head>

<div class="centered-container">
  <div class="box-container">
    <h1 class="page-title">Mechanus</h1>
    <a href="./devices" class="action-button">Devices</a>
    <a href="./users/login" class="action-button">Login</a>
    <a href="./users/signup" class="action-button">Signup</a>
  </div>
</div>
