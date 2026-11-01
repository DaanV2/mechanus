<script lang="ts">
  import { page } from '$app/state';
  import { ScreenRenderer } from '$lib/2d/screen-renderer';
  import { deviceId, randomId, type HardwareType } from '$lib/devices/identity';
  import { WebsocketHandler } from '$lib/networking/websocket';
  import { create } from '@bufbuild/protobuf';
  import { onMount } from 'svelte';
  import { InitialSetupRequestSchema } from '../../../proto/screens/v1/setup_pb';

  const RECONNECT_DELAY_MS = 3000;

  let container: HTMLDivElement;
  let status = $state<'connecting' | 'connected' | 'disconnected'>('connecting');
  let screenId = $state('');

  onMount(() => {
    // `?device=<id>` overrides the persisted device ID, `?type=<hardware>` reports the hardware type
    screenId = page.url.searchParams.get('device') || deviceId();
    const hardware = (page.url.searchParams.get('type') || 'screen') as HardwareType;
    const connectionId = randomId();

    const renderer = new ScreenRenderer();
    let conn: WebsocketHandler | undefined;
    let reconnect: ReturnType<typeof setTimeout> | undefined;
    let stopped = false;

    function connect() {
      status = 'connecting';
      conn = new WebsocketHandler(screenId, connectionId, undefined, {
        deviceId: screenId,
        deviceType: hardware
      });
      conn.events.add('initialSetup', (_, msg) => renderer.setState(msg.state));
      conn.events.add('screenState', (_, msg) => renderer.setState(msg));
      conn.events.add('splashScreenUpdate', (_, msg) => renderer.setSplash(msg));
      conn.addEventListener('open', function () {
        status = 'connected';
        this.send({
          action: { case: 'initialSetupRequest', value: create(InitialSetupRequestSchema, {}) }
        });
      });
      conn.addEventListener('close', () => {
        status = 'disconnected';
        if (!stopped) reconnect = setTimeout(connect, RECONNECT_DELAY_MS);
      });
    }

    renderer.init(container).then(connect);

    return () => {
      stopped = true;
      clearTimeout(reconnect);
      conn?.socket.close();
      renderer.destroy();
    };
  });
</script>

<svelte:head>
  <title>Mechanus screen</title>
</svelte:head>

<div class="fixed inset-0 bg-black" bind:this={container}></div>

{#if status !== 'connected'}
  <div class="fixed right-2 bottom-2 rounded bg-gray-800/80 px-3 py-1 text-sm text-gray-200">
    {status === 'connecting' ? 'Connecting' : 'Disconnected, retrying'} · device {screenId}
  </div>
{/if}
