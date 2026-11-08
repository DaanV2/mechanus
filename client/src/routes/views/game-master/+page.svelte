<script lang="ts">
  import { goto } from '$app/navigation';
  import SimpleCard from '$lib/components/cards/simple.svelte';
  import Fluid from '$lib/components/collections/fluid.svelte';
  import ErrorMessage from '$lib/components/error-message.svelte';
  import type { MechanusError } from '$lib/components/errors';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import { createSceneClient } from '$lib/stores/clients';
  import {
    Badge,
    Button,
    Checkbox,
    Heading,
    Hr,
    Input,
    Label,
    Modal,
    Select,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell
  } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import {
    EntityKind,
    type ActiveScene,
    type Entity,
    type Scene
  } from '../../../proto/scenes/v1/scenes_pb';

  const KINDS = [
    { value: EntityKind.ENEMY, name: 'Enemy' },
    { value: EntityKind.NPC, name: 'NPC' },
    { value: EntityKind.PLAYER, name: 'Player' },
    { value: EntityKind.OBJECT, name: 'Object' }
  ];
  const kindName = (kind: EntityKind) => KINDS.find((k) => k.value === kind)?.name ?? 'Unknown';

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
  const setHidden = (entity: Entity, hidden: boolean) =>
    run(() => createSceneClient().setEntityHidden({ entityId: entity.id, hidden }));
  const remove = (entity: Entity) =>
    run(() => createSceneClient().removeEntity({ entityId: entity.id }));

  // Reset: throws away what happened in a scene during play, asked for in a modal first
  let resetScene = $state<Scene | undefined>(undefined);
  let resetOpen = $state(false);
  const askReset = (scene: Scene) => ((resetScene = scene), (resetOpen = true));

  async function reset() {
    const sceneId = resetScene?.id ?? '';
    resetOpen = false;
    await run(() => createSceneClient().reset({ sceneId }));
  }

  // Spawn form: a new entity, by default hidden so the GM can place it before revealing it
  let spawnId = $state('');
  let spawnName = $state('');
  let spawnKind = $state(EntityKind.ENEMY);
  let spawnHidden = $state(true);
  let spawnX = $state(0);
  let spawnY = $state(0);

  async function spawn(event: SubmitEvent) {
    event.preventDefault();
    const entityId = spawnId.trim();
    await run(() =>
      createSceneClient().spawnEntity({
        entityId,
        name: spawnName.trim() || entityId,
        kind: Number(spawnKind),
        hidden: spawnHidden,
        position: { x: Number(spawnX), y: Number(spawnY) }
      })
    );
    if (!errorObj) {
      spawnId = '';
      spawnName = '';
    }
  }

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
          <Button
            size="sm"
            color="red"
            outline
            class="ms-auto"
            disabled={busy}
            onclick={() => active?.scene && askReset(active.scene)}>Reset scene</Button
          >
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

    <div class="w-full max-w-3xl px-5 py-3">
      <Heading tag="h3" class="py-2 text-xl">Entities</Heading>
      <Table>
        <TableHead>
          <TableHeadCell>Name</TableHeadCell>
          <TableHeadCell>Kind</TableHeadCell>
          <TableHeadCell>Status</TableHeadCell>
          <TableHeadCell><span class="sr-only">Actions</span></TableHeadCell>
        </TableHead>
        <TableBody>
          {#each active.entities as entity (entity.id)}
            <TableBodyRow data-testid={`entity-${entity.id}`}>
              <TableBodyCell>{entity.name || entity.id}</TableBodyCell>
              <TableBodyCell>{kindName(entity.kind)}</TableBodyCell>
              <TableBodyCell>
                <Badge color={entity.hidden ? 'gray' : 'green'}>
                  {entity.hidden ? 'Hidden' : 'Visible'}
                </Badge>
              </TableBodyCell>
              <TableBodyCell class="space-x-1 text-right">
                {#if entity.hidden}
                  <Button size="xs" disabled={busy} onclick={() => setHidden(entity, false)}
                    >Reveal</Button
                  >
                {:else}
                  <Button
                    size="xs"
                    color="alternative"
                    disabled={busy}
                    onclick={() => setHidden(entity, true)}>Hide</Button
                  >
                {/if}
                <Button size="xs" color="red" disabled={busy} onclick={() => remove(entity)}
                  >Remove</Button
                >
              </TableBodyCell>
            </TableBodyRow>
          {:else}
            <TableBodyRow>
              <TableBodyCell colspan={4}>No entities in this scene</TableBodyCell>
            </TableBodyRow>
          {/each}
        </TableBody>
      </Table>

      <form class="mt-3 flex flex-wrap items-end gap-2" aria-label="Spawn entity" onsubmit={spawn}>
        <Label class="space-y-1">
          <span>ID</span>
          <Input
            type="text"
            size="sm"
            required
            maxlength={64}
            placeholder="wolf-1"
            bind:value={spawnId}
          />
        </Label>
        <Label class="space-y-1">
          <span>Name</span>
          <Input type="text" size="sm" maxlength={64} placeholder="Wolf" bind:value={spawnName} />
        </Label>
        <Label class="space-y-1">
          <span>Kind</span>
          <Select size="sm" items={KINDS} bind:value={spawnKind} />
        </Label>
        <Label class="w-20 space-y-1">
          <span>X</span>
          <Input type="number" size="sm" step="any" bind:value={spawnX} />
        </Label>
        <Label class="w-20 space-y-1">
          <span>Y</span>
          <Input type="number" size="sm" step="any" bind:value={spawnY} />
        </Label>
        <Checkbox bind:checked={spawnHidden}>Hidden</Checkbox>
        <Button type="submit" size="sm" disabled={busy}>Spawn</Button>
      </form>
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
              <Button size="sm" color="alternative" disabled={busy} onclick={() => askReset(scene)}
                >Reset</Button
              >
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

<Modal title="Reset scene" bind:open={resetOpen} size="sm">
  <p>
    Reset <strong>{resetScene?.name || resetScene?.id}</strong> to how it is configured? Spawned entities
    are gone, moved, hidden and removed entities go back to their start, live adjustments and cameras
    are cleared and the default variant is active. This can't be undone.
  </p>
  <div class="flex gap-2">
    <Button color="red" onclick={reset}>Reset</Button>
    <Button color="alternative" onclick={() => (resetOpen = false)}>Cancel</Button>
  </div>
</Modal>
