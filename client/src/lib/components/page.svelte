<script lang="ts">
  import type { Snippet } from 'svelte';

  /**
   * The main content area of a page: centred, with side gutters, and growing so the footer below it
   * stays at the bottom of short pages. `width` picks the maximum content width.
   */
  let {
    width = 'default',
    title = undefined,
    class: className = '',
    children
  }: {
    width?: 'narrow' | 'default' | 'wide';
    title?: string;
    class?: string;
    children?: Snippet;
  } = $props();

  const widths = {
    narrow: 'max-w-md',
    default: 'max-w-3xl',
    wide: 'max-w-6xl'
  };
</script>

<main class="mx-auto w-full grow px-4 py-6 sm:px-6 lg:px-8 {widths[width]} {className}">
  {#if title}
    <h1 class="text-ink mb-4 text-2xl font-bold">{title}</h1>
  {/if}
  {@render children?.()}
</main>
