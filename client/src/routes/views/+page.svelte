<script lang="ts">
  import { userHandler, UserState } from '$lib/handlers/user';
  import { Card } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';

  let userData: UserState = $state(UserState.LOGGED_OUT);

  onMount(() => {
    userData = userHandler.current;
  });
</script>

<svelte:head>
  <title>Mechanus</title>
</svelte:head>

<NavBar />

<div class="flex min-h-screen flex-col items-center py-5">
  <!-- User Data -->
  {#if userData}
    <Card href="./game-master" class="m-5 p-4 sm:p-6 md:p-8">
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
        Game Master
      </h5>
    </Card>
    <Card href="./players" class="m-5 p-4 sm:p-6 md:p-8">
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Player</h5>
    </Card>
  {/if}
  <Card href="./devices" class="m-5 p-4 sm:p-6 md:p-8">
    <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Devices</h5>
  </Card>
</div>

<Footer />
