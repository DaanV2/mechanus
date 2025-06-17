<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { UserHandler } from '../../../lib/handlers/user';
  import { User } from '../../../proto/users/v1/users_pb';
  import NavBar from '../../../components/nav-bar.svelte';
  import Footer from '../../../components/footer.svelte';

  let user = $state<User | undefined>(undefined);

  onMount(async () => {
    const handler = UserHandler.instance();

    if (!handler.hasLoggedinUser) {
      // redirect to login
      return goto('/users/login');
    }

    const data = await handler.serverData();
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
