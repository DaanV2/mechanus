<script lang="ts">
  import { goto } from '$app/navigation';
  import SimpleCard from '$lib/components/cards/simple.svelte';
  import Fluid from '$lib/components/collections/fluid.svelte';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import { createSceneClient } from '$lib/stores/clients';
  import { Badge, Button, Heading, Hr } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import type { ActiveScene, Scene } from '../../../proto/scenes/v1/scenes_pb';

  let scenes = $state<Scene[] | undefined>(undefined);
  let active = $state<ActiveScene | undefined>(undefined);
  let deviceErrors = $state<string[]>([]);
  let busy = $state(false);
  let errorObj = $state<MechanusError>(null);

  async function load() {
    const client = createSceneClient();
    const [list, current] = await Promise.all([client.list({}), client.getActive({})]);
    scenes = list.scenes;
    active = current.active;
  }

  async function run(action: () => Promise<{ active?: ActiveScene; deviceErrors: string[] }>) {
    errorObj = null;
    busy = true;
    try {
      const resp = await action();
      active = resp.active;
      deviceErrors = resp.deviceErrors;
    } catch (err) {
      errorObj = err instanceof Error ? err : new Error('An unexpected error occurred.');
    } finally {
      busy = false;
    }
  }

  const activate = (scene: Scene) => run(() => createSceneClient().activate({ sceneId: scene.id }));
  const setVariant = (variantId: string) =>
    run(() => createSceneClient().setVariant({ variantId }));

  onMount(async () => {
    if (!userHandler.current.loggedin) {
      return goto('/users/login');
    }

    try {
      await load();
    } catch (err) {
      errorObj = err instanceof Error ? err : new Error('An unexpected error occurred.');
    }
  });
</script>

<svelte:head>
  <title>Game master</title>
</svelte:head>

<NavBar />

<Heading tag="h2" class="px-5 py-3">Scenes</Heading>
<Hr />

<div class="flex flex-col items-center">
  <ErrorMessage title="Scenes" error={errorObj} />

  {#if active?.scene}
    <div class="w-full max-w-3xl px-5">
      <SimpleCard title={`Active: ${active.scene.name || active.scene.id}`}>
        <div class="flex flex-wrap gap-2">
          <Button
            size="sm"
            color={active.variantId === '' ? 'primary' : 'alternative'}
            disabled={busy}
            onclick={() => setVariant('')}>Base</Button
          >
          {#each active.scene.variants as variant (variant.id)}
            <Button
              size="sm"
              color={active.variantId === variant.id ? 'primary' : 'alternative'}
              disabled={busy}
              onclick={() => setVariant(variant.id)}>{variant.name || variant.id}</Button
            >
          {/each}
        </div>
        {#if deviceErrors.length > 0}
          <ul class="mt-3 text-sm text-red-600 dark:text-red-400">
            {#each deviceErrors as deviceError, i (i)}
              <li>{deviceError}</li>
            {/each}
          </ul>
        {/if}
      </SimpleCard>
    </div>
  {/if}

  {#if scenes === undefined}
    <p class="p-4 text-gray-500">Loading scenes...</p>
  {:else if scenes.length === 0}
    <p class="p-4 text-gray-500">
      No scenes configured yet. Add scene files to the <code>scenes</code> folder of the server's config
      folder.
    </p>
  {:else}
    <Fluid>
      {#each scenes as scene (scene.id)}
        <SimpleCard title={scene.name || scene.id} text={scene.description}>
          <div class="mt-3 flex items-center gap-2">
            {#if active?.scene?.id === scene.id}
              <Badge color="green">Active</Badge>
            {:else}
              <Button size="sm" disabled={busy} onclick={() => activate(scene)}>Activate</Button>
            {/if}
            {#if scene.variants.length > 0}
              <span class="text-sm text-gray-500">{scene.variants.length} variant(s)</span>
            {/if}
          </div>
        </SimpleCard>
      {/each}
    </Fluid>
  {/if}
</div>
