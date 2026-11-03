import { createDeviceClient } from '../stores/clients';

export class DeviceHandler {
  /** Pairs the device showing the code under the given name, requires the admin role */
  async confirm(code: string, name: string) {
    const resp = await createDeviceClient().confirm({ code, name });

    return resp.device;
  }

  /** Lists the paired and connected devices, requires the admin role */
  async list() {
    const resp = await createDeviceClient().list({});

    return resp.devices;
  }

  /** Un-pairs a device, it drops back to the pairing page. Requires the admin role */
  async revoke(id: string) {
    await createDeviceClient().revoke({ id });
  }

  /** Binds a paired device to a target of the room, an empty target removes the binding. Requires the admin role */
  async bind(id: string, target: string) {
    const resp = await createDeviceClient().bind({ id, target });

    return resp.device;
  }
}

export const deviceHandler = new DeviceHandler();
