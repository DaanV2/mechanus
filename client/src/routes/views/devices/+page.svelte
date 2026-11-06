<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { ScreenRenderer } from '$lib/2d/screen-renderer';
  import { Session } from '$lib/authentication/session';
  import { DeviceCredentials } from '$lib/devices/credentials';
  import { deviceId, randomId, type HardwareType } from '$lib/devices/identity';
  import { WebsocketHandler } from '$lib/networking/websocket';
  import { create } from '@bufbuild/protobuf';
  import { onMount } from 'svelte';
  import { MoveEntitySchema, SetCameraSchema } from '../../../proto/screens/v1/play_actions_pb';
  import { InitialSetupRequestSchema } from '../../../proto/screens/v1/setup_pb';
  import { CameraSchema, Vector2Schema } from '../../../proto/screens/v1/state_pb';

  const RECONNECT_DELAY_MS = 3000;
  /** WebSocket close status the server uses when the credential of a connection expired or got revoked */
  const STATUS_POLICY_VIOLATION = 1008;

  let container: HTMLDivElement;
  let status = $state<'connecting' | 'connected' | 'disconnected'>('connecting');
  let screenId = $state('');

  onMount(() => {
    // `?device=<id>` overrides the persisted device ID, `?type=<hardware>` reports the hardware type
    const override = page.url.searchParams.get('device');
    screenId = override || deviceId();
    const hardware = (page.url.searchParams.get('type') || 'screen') as HardwareType;
    const connectionId = randomId();

    // A paired device connects as itself. Looking at another device's screen takes a logged in user
    const asDevice = !override && DeviceCredentials.canRefresh();
    const hasCredentials = () =>
      asDevice ? DeviceCredentials.canRefresh() : !!Session.claims() || Session.canRefresh();
    const toPairing = () => goto(`/devices/pair/?type=${encodeURIComponent(hardware)}`);

    if (!hasCredentials()) {
      toPairing();

      return;
    }

    const renderer = new ScreenRenderer();
    let conn: WebsocketHandler | undefined;
    let reconnect: ReturnType<typeof setTimeout> | undefined;
    let stopped = false;

    // Reconnects refresh the access token first, the server closes the connection once the token expires
    async function connect() {
      status = 'connecting';
      const query = { deviceId: screenId, deviceType: hardware };
      conn = asDevice
        ? await WebsocketHandler.openAsDevice(screenId, connectionId, query)
        : await WebsocketHandler.open(screenId, connectionId, query);
      if (stopped) return conn.socket.close();
      let opened = false;
      conn.events.add('initialSetup', (_, msg) => renderer.setState(msg.state));
      conn.events.add('screenState', (_, msg) => renderer.setState(msg));
      conn.events.add('splashScreenUpdate', (_, msg) => renderer.setSplash(msg));
      conn.events.add('entityUpdate', (_, msg) => renderer.updateEntity(msg));
      conn.events.add('entityRemove', (_, msg) => renderer.removeEntity(msg));
      conn.events.add('cameraUpdate', (_, msg) => renderer.setCamera(msg));
      conn.events.add('actionError', (id, msg) =>
        console.warn('action refused', id, msg.code, msg.message)
      );
      conn.addEventListener('open', function () {
        opened = true;
        status = 'connected';
        this.send({
          action: { case: 'initialSetupRequest', value: create(InitialSetupRequestSchema, {}) }
        });
      });
      conn.addEventListener('close', async (ev) => {
        status = 'disconnected';
        if (stopped) return;

        // A browser can't see why a connection got refused. A refresh tells: the server refusing it forgets the
        // tokens, and a revoked device goes back to pairing. A network error keeps them, and keeps retrying
        if (!opened || ev.code === STATUS_POLICY_VIOLATION) {
          await (asDevice ? DeviceCredentials.refresh() : Session.refresh());
          if (stopped) return;
          if (!hasCredentials()) {
            toPairing();

            return;
          }
        }

        reconnect = setTimeout(connect, RECONNECT_DELAY_MS);
      });
    }

    // A logged in user looking at a screen can move tokens and the camera. A paired device only shows it: touch
    // control is configured per device, which is not built yet
    if (!asDevice) {
      renderer.enableControls({
        moveEntity: (entityId, position) =>
          conn?.send({
            action: {
              case: 'moveEntity',
              value: create(MoveEntitySchema, {
                entityId,
                position: create(Vector2Schema, position)
              })
            }
          }),
        setCamera: (center, zoom) =>
          conn?.send({
            action: {
              case: 'setCamera',
              value: create(SetCameraSchema, {
                camera: create(CameraSchema, { center: create(Vector2Schema, center), zoom })
              })
            }
          })
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
