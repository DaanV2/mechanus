<script lang="ts">
  import { goto } from '$app/navigation';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import { onMount } from 'svelte';
  import type { User } from '../../../proto/users/v1/users_pb';

  let user = $state<User | undefined>(undefined);

  onMount(async () => {
    console.log('document.cookie:', document.cookie);
    console.log('userHandler.current:', userHandler.current);

    if (!userHandler.current.loggedin) {
      console.error('not logged in');
      // redirect to login
      return goto('/users/login');
    }

    const data = await userHandler.serverData();
    user = data.user;
  });
</script>

<svelte:head>
  {#if user}
    <title>User - {user.name}</title>
  {:else}
    <title>User</title>
  {/if}
</svelte:head>

<NavBar />

<div class="flex min-h-screen flex-col items-center justify-center">
  <div class="w-full max-w-md space-y-4 rounded-lg border border-gray-700 bg-gray-800 p-8 shadow-lg">
    <h1 class="text-2xl font-bold text-white">Welcome{user ? `, ${user.name}` : ''}!</h1>
    {#if user}
      <div class="space-y-2">
        <p class="text-sm text-gray-400">Username</p>
        <p class="font-medium text-white">{user.name}</p>
      </div>
    {/if}
  </div>
</div>

<Footer />
