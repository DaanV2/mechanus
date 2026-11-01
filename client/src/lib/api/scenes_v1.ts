import { createClient, type Client, type Transport } from '@connectrpc/connect';
import { SceneService } from '../../proto/scenes/v1/scenes_pb';

export function createSceneClient(transport: Transport): Client<typeof SceneService> {
  return createClient(SceneService, transport);
}
