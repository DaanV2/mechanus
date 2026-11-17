<script lang="ts">
  import Page from '$lib/components/page.svelte';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { Heading } from 'flowbite-svelte';
  import Simple from '../../lib/components/cards/simple.svelte';
  import Fluid from '../../lib/components/collections/fluid.svelte';

  // Define an array of subpage objects
  const subpages = [
    { name: 'Size check', path: '/demo/size' },
    { name: 'Playground', path: '/demo/playground' }
  ];
</script>

<svelte:head>
  <title>Mechanus - Demos & Tools</title>
</svelte:head>

<NavBar />

<Page width="wide">
  <Heading class="text-ink pb-3" tag="h2">Demos and Tools</Heading>

  <Fluid>
    {#each subpages as { name, path } (name)}
      <Simple href={path} title={name}></Simple>
    {/each}
  </Fluid>
</Page>

<Footer />
