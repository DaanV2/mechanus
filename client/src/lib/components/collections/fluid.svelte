<script lang="ts">
  import type { Snippet } from 'svelte';

  let { children }: { children?: Snippet } = $props();
</script>

<!-- Wrapping, centred row of cards -->
<div class="flex w-full flex-wrap items-stretch justify-center py-3">
  {@render children?.()}
</div>
