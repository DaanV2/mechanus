<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { Application } from '$lib/2d/application';

  const app = new Application();

  onMount(async () => {
    await app.init();
  });
  onDestroy(() => app.destroy());
</script>
