<script lang="ts">
  import SimpleCard from '$lib/components/cards/simple.svelte';
  import Fluid from '$lib/components/collections/fluid.svelte';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler, UserState } from '$lib/handlers/user';
  import { Heading, Hr } from 'flowbite-svelte';
  import { onMount } from 'svelte';

  let userData: UserState = $state(UserState.LOGGED_OUT);

  onMount(() => {
    userData = userHandler.current;
  });
</script>

<svelte:head>
  <title>Mechanus</title>
</svelte:head>

<NavBar />

<Heading tag="h2" class="px-5 py-3">Choose your adventure</Heading>
<Hr />

<Fluid>
  <!-- User Data -->
  {#if userData}
    <SimpleCard href="/campaigns" title="Campaigns" text="see your campaigns" />
    <SimpleCard href="/users/profile" title="Profile" text="Manage your user Profile" />
    <SimpleCard href="/devices" title="Devices" text="Manage the devices" />
  {/if}
  <!-- Aviable for all -->
  <SimpleCard href="/views" title="Views" text="..." />
  <SimpleCard
    href="/demo"
    title="Demos & Tools"
    text="A bunch of demos and tools to ensure everything is working"
  />
</Fluid>

<Footer />
