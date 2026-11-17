<script>
  import { Card } from 'flowbite-svelte';
  let { title = undefined, text = undefined, href = undefined, children = undefined } = $props();
</script>

<Card
  {href}
  class="bg-surface border-line dark:bg-surface dark:border-line m-2 p-4 sm:p-6 md:p-8 {href
    ? 'hover:border-primary dark:hover:border-primary transition-[border-color,box-shadow] hover:shadow-lg'
    : ''}"
>
  {#if title && title.length > 0}
    <h5 class="text-ink mb-2 text-2xl font-bold tracking-tight">{title}</h5>
  {/if}
  {#if text && text.length > 0}
    <p class="text-muted leading-tight font-normal">{text}</p>
  {/if}
  {@render children?.()}
</Card>
