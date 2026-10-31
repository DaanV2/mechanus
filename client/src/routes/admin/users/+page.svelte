<script lang="ts">
  import { goto } from '$app/navigation';
  import Footer from '$lib/components/footer.svelte';
  import NavBar from '$lib/components/nav-bar.svelte';
  import { userHandler } from '$lib/handlers/user';
  import {
    Badge,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell
  } from 'flowbite-svelte';
  import { onMount } from 'svelte';
  import type { User } from '../../../proto/users/v1/users_pb';

  let users = $state<User[] | undefined>(undefined);
  let error = $state<string | undefined>(undefined);

  onMount(async () => {
    if (!userHandler.current.loggedin) {
      return goto('/users/login');
    }
    if (!userHandler.isAdmin) {
      error = 'You need to be an admin to view this page';

      return;
    }

    try {
      const result = await userHandler.list();
      users = result.sort((a, b) => a.name.localeCompare(b.name));
    } catch (err) {
      console.error('failed to list users', err);
      error = err instanceof Error ? err.message : String(err);
    }
  });
</script>

<svelte:head>
  <title>Admin - Users</title>
</svelte:head>

<NavBar />

<div class="flex min-h-screen flex-col items-center p-8">
  <div class="w-full max-w-4xl space-y-4">
    <h1 class="text-2xl font-bold text-white">Users</h1>

    {#if error}
      <div role="alert" class="rounded-lg bg-red-900 p-4 text-white">{error}</div>
    {:else if users === undefined}
      <p class="text-gray-400">Loading users...</p>
    {:else}
      <p class="text-sm text-gray-400">{users.length} users</p>
      <Table aria-label="Users">
        <TableHead>
          <TableHeadCell>Username</TableHeadCell>
          <TableHeadCell>Roles</TableHeadCell>
          <TableHeadCell>ID</TableHeadCell>
        </TableHead>
        <TableBody>
          {#each users as user (user.id)}
            <TableBodyRow>
              <TableBodyCell>{user.name}</TableBodyCell>
              <TableBodyCell>
                {#each user.roles as role (role)}
                  <Badge class="me-1">{role}</Badge>
                {/each}
              </TableBodyCell>
              <TableBodyCell class="font-mono text-xs">{user.id}</TableBodyCell>
            </TableBodyRow>
          {/each}
        </TableBody>
      </Table>
    {/if}
  </div>
</div>

<Footer />
