<script lang="ts">
  import { Application } from '$lib/2d/application';
  import { WebsocketHandler } from '$lib/networking/websocket';
  import { onMount } from 'svelte';

  const app = new Application();

  onMount(async () => {
    const conn = new WebsocketHandler('players', '');
    await app.init(conn);
  });
</script>
