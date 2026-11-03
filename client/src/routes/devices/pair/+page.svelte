<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { DeviceCredentials } from '$lib/devices/credentials';
  import { deviceId, type HardwareType } from '$lib/devices/identity';
  import { Code, ConnectError } from '@connectrpc/connect';
  import { onMount } from 'svelte';

  const RETRY_DELAY_MS = 3000;

  let code = $state('');
  let id = $state('');
  let status = $state<'connecting' | 'showing' | 'replaced'>('connecting');

  /** Runs one pairing stream, returns once the device is paired or the stream ended */
  async function pair(hardware: HardwareType, signal: AbortSignal): Promise<boolean> {
    for await (const resp of DeviceCredentials.pair(hardware, signal)) {
      switch (resp.event.case) {
        case 'code':
          code = resp.event.value.code;
          status = 'showing';
          break;
        case 'tokens':
          DeviceCredentials.store(resp.event.value);

          return true;
      }
    }

    return false;
  }

  onMount(() => {
    const hardware = (page.url.searchParams.get('type') || 'screen') as HardwareType;
    const view = `/views/devices/?type=${encodeURIComponent(hardware)}`;
    if (DeviceCredentials.canRefresh()) {
      goto(view);

      return;
    }

    id = deviceId();
    const abort = new AbortController();

    // A code is only valid while the stream is open: when it breaks, show nothing until there is a new one
    async function run() {
      while (!abort.signal.aborted) {
        try {
          if (await pair(hardware, abort.signal)) {
            await goto(view);

            return;
          }
        } catch (err) {
          if (abort.signal.aborted) return;
          console.error('pairing stream ended', err);
          if (err instanceof ConnectError && err.code === Code.Aborted) {
            code = '';
            status = 'replaced';

            return;
          }
        }

        code = '';
        status = 'connecting';
        await new Promise((resolve) => setTimeout(resolve, RETRY_DELAY_MS));
      }
    }

    run();

    return () => abort.abort();
  });
</script>

<svelte:head>
  <title>Pair this device</title>
</svelte:head>

<div class="fixed inset-0 flex flex-col items-center justify-center gap-8 bg-black p-8 text-center">
  {#if status === 'showing'}
    <p class="text-2xl text-gray-300">Enter this code on the devices page to pair this screen</p>
    <p
      class="font-mono text-[20vw] leading-none font-bold tracking-[0.2em] text-white"
      data-testid="pairing-code"
    >
      {code}
    </p>
  {:else if status === 'replaced'}
    <p class="text-2xl text-gray-300">This device started pairing in another window.</p>
  {:else}
    <p class="text-2xl text-gray-300">Connecting to the server…</p>
  {/if}
  <p class="text-sm text-gray-500">Device {id}</p>
</div>
