<script lang="ts">
  import { Card, Heading } from 'flowbite-svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import Footer from '$lib/components/footer.svelte';
  import Fluid from '../../lib/components/collections/fluid.svelte';
  import Simple from '../../lib/components/cards/simple.svelte';

  // Define an array of subpage objects
  const subpages = [{ name: 'Size check', path: '/demo/size' }];
</script>

<svelte:head>
  <title>Mechanus - Demos & Tools</title>
</svelte:head>

<NavBar />

<Heading class="px-3 py-4" tag="h2">Demos and Tools</Heading>

<Fluid>
  {#each subpages as { name, path } (name)}
    <Simple href={path} title={name}></Simple>
  {/each}
</Fluid>

<Footer />
