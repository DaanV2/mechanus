import { createClient, type Client, type Transport } from '@connectrpc/connect';
import { DeviceService } from '../../proto/devices/v1/devices_pb';

export function createDeviceClient(transport: Transport): Client<typeof DeviceService> {
  return createClient(DeviceService, transport);
}
