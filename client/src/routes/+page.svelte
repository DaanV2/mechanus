<script lang="ts">
  import { Card } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import Footer from '../components/footer.svelte';
  import NavBar from '../components/nav-bar.svelte';
  import type { JWTClaims } from '../lib/authenication/jwt/parse';
  import { UserHandler } from '../lib/handlers/user';

  let user: UserHandler | undefined = $state(undefined);
  let userData: JWTClaims | undefined = $state(undefined);

  onMount(() => {
    user = UserHandler.instance();
    userData = user.data();
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
    <Card href="/profile" class="m-5 p-4 sm:p-6 md:p-8">
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
