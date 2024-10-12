import { error } from '@sveltejs/kit';

export function load({params}) {
    if (typeof params.playerId !== "string") {
        error(400, "Page not found")
    }

    return {
        playerId: params.playerId,
    }
}