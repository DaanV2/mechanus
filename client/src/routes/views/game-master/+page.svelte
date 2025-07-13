<script lang="ts">
  import { Application } from '$lib/2d/application';
  import { onMount } from 'svelte';

  const app = new Application();

  onMount(async () => {
    await app.init();
  });
</script>
