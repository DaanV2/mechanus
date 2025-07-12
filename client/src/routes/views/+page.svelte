<script lang="ts">
  import { userHandler, UserState } from '$lib/handlers/user';
  import { Card } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import SimpleCard from '$lib/components/cards/simple.svelte';
  import Fluid from '$lib/components/collections/fluid.svelte';

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
  <Fluid>
    {#if userData}
      <SimpleCard href="/views/game-master" title="Game Master" />
      <SimpleCard href="/views/players" title="Player" />
    {/if}
    <SimpleCard href="/views/devices" title="Devices" />
  </Fluid>
</div>

<Footer />
