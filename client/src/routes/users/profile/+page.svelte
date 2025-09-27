<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
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
      return goto(resolve('/users/login', {}));
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

<div class="flex flex-col items-center py-5">
  <form>
    <h1>Welcome {user?.name}!</h1>
    <div class="box-container">
      {#if user}
        <p id="user.name">name: {user.name}</p>
      {/if}
    </div>
  </form>
</div>

<Footer />
