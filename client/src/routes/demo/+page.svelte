<script lang="ts">
  import { Card } from 'flowbite-svelte';
  import NavBar from '../../components/nav-bar.svelte';
  import Footer from '../../components/footer.svelte';

  // Define an array of subpage objects
  const subpages = [{ name: 'Size check', path: '/demo/size' }];
</script>

<svelte:head>
  <title>Mechanus - Demos & Tools</title>
</svelte:head>

<NavBar />

<div class="flex min-h-screen flex-col items-center py-5">
  {#each subpages as { name, path }}
    <Card href={path} class="m-5 p-4 sm:p-6 md:p-8">
      <h3 class="font-normal leading-tight text-gray-700 dark:text-gray-400">{name}</h3>
    </Card>
  {/each}
</div>

<Footer />
