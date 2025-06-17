<script>
  import { Blockquote, Card, Heading, Hr } from 'flowbite-svelte';
  import Footer from '../../components/footer.svelte';
  import NavBar from '../../components/nav-bar.svelte';
</script>

<svelte:head>
  <title>Mechanus - About</title>
</svelte:head>

<NavBar />

<div class="flex min-h-screen flex-col items-center py-5">
  <Heading tag="h2" class="py-2 text-white dark:text-white">About</Heading>
  <Blockquote class="p-4 sm:p-6 md:p-8">
    <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">TODO</p>
  </Blockquote>

  <Hr class="mx-auto my-4 h-1 w-48 rounded md:my-10" />
  <Heading tag="h2" class="py-2 text-white dark:text-white">Licences</Heading>
  <Card class="p-4 sm:p-6 md:p-8"><p>TODO</p></Card>
</div>

<Footer />
