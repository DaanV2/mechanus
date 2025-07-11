<script lang="ts">
  import { Application } from '$lib/2d/application';
  import { onDestroy, onMount } from 'svelte';

  const app = new Application();

  onMount(async () => {
    await app.init();
  });
  onDestroy(() => app.destroy());
</script>
