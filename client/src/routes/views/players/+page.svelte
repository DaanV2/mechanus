<script lang="ts">
  import { goto } from '$app/navigation';
  import { Application } from '$lib/2d/application';
  import { randomId } from '$lib/devices/identity';
  import { userHandler } from '$lib/handlers/user';
  import { WebsocketHandler } from '$lib/networking/websocket';
  import { create } from '@bufbuild/protobuf';
  import { onMount } from 'svelte';
  import { InitialSetupRequestSchema } from '../../../proto/screens/v1/setup_pb';

  const RECONNECT_DELAY_MS = 3000;

  const app = new Application();

  onMount(() => {
    const connectionId = randomId();
    let conn: WebsocketHandler | undefined;
    let reconnect: ReturnType<typeof setTimeout> | undefined;
    let stopped = false;

    // Reconnects refresh the access token first, the server closes the connection once the token expires
    async function connect() {
      conn = await WebsocketHandler.open('players', connectionId);
      if (stopped) return conn.socket.close();

      conn.addEventListener('open', function () {
        this.send({
          action: { case: 'initialSetupRequest', value: create(InitialSetupRequestSchema, {}) }
        });
      });
      conn.addEventListener('close', () => {
        if (!stopped) reconnect = setTimeout(connect, RECONNECT_DELAY_MS);
      });
    }

    userHandler.restore().then(async (user) => {
      if (!user.loggedin) return goto('/users/login');

      await app.init();
      await connect();
    });

    return () => {
      stopped = true;
      clearTimeout(reconnect);
      conn?.socket.close();
    };
  });
</script>
