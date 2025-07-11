<script lang="ts">
  import { userHandler, UserState } from '$lib/handlers/user';
  import { Card } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import Footer from '../lib/components/footer.svelte';
  import NavBar from '../lib/components/nav-bar.svelte';

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
    <Card href="/campaigns" class="m-5 p-4 sm:p-6 md:p-8">
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
        Campaigns
      </h5>
      <p class="font-normal leading-tight text-gray-700 dark:text-gray-400">see your campaigns</p>
    </Card>
    <Card href="/users/profile" class="m-5 p-4 sm:p-6 md:p-8">
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Profile</h5>
      <p class="font-normal leading-tight text-gray-700 dark:text-gray-400">
        Manage your user Profile
      </p>
    </Card>
  {/if}
  <!-- Aviable for all -->
  <Card href="/devices" class="m-5 p-4 sm:p-6 md:p-8">
    <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Devices</h5>
    <p class="font-normal leading-tight text-gray-700 dark:text-gray-400">Manage the devices</p>
  </Card>
  <Card href="/demo" class="m-5 p-4 sm:p-6 md:p-8">
    <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
      Demos & Tools
    </h5>
    <p class="font-normal leading-tight text-gray-700 dark:text-gray-400">
      A bunch of demos and tools to ensure everything is working
    </p>
  </Card>
</div>

<Footer />
