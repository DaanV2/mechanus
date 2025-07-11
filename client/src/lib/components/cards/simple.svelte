<script>
  import { Card } from 'flowbite-svelte';
  let { title = undefined, text = undefined, href = undefined, children = undefined } = $props();
</script>

<Card {href} class="m-2 p-4 sm:p-6 md:p-8">
  {#if title && title.length > 0}
    <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">{title}</h5>
  {/if}
  {#if text && text.length > 0}
    <p class="font-normal leading-tight text-gray-700 dark:text-gray-400">{text}</p>
  {/if}
  {@render children?.()}
</Card>
