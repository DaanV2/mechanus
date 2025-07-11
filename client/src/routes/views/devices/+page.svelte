<script lang="ts">
  import { onMount } from 'svelte';
  import { Application } from '$lib/2d/application';

  const app = new Application();

  onMount(async () => {
    await app.init();
  });
</script>
