<script lang="ts">
  import { Application } from '$lib/2d/application';
  import { onMount } from 'svelte';

  const app = new Application();

  onMount(async () => {
    await app.init();
    // TODO add your pixi test code here, but don't commit :)
  });
</script>
