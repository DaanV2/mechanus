import { error } from '@sveltejs/kit';

export function load({params}) {
    if (typeof params.deviceId !== "string") {
        error(400, "Page not found")
    }

    return {
        deviceId: params.deviceId,
    }
}