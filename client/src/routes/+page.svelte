<script lang="ts">
  import { onMount } from 'svelte';
  import NavBar from '../components/nav-bar.svelte';
  import Footer from '../components/footer.svelte';
  import { Card } from 'flowbite-svelte';

  onMount(() => {
    console.log('cookies', document.cookie);
  });
</script>

<svelte:head>
  <title>Mechanus</title>
</svelte:head>

<NavBar />

<div class="flex min-h-screen flex-col items-center py-5">
  <Card href="/devices" class="m-5 p-4 sm:p-6 md:p-8">
    <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Devices</h5>
    <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">Manage the devices</p>
  </Card>
  <Card href="/demo" class="m-5 p-4 sm:p-6 md:p-8">
    <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
      Demos & Tools
    </h5>
    <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
      A bunch of demos and tools to ensure everything is working
    </p>
  </Card>
</div>

<Footer />
