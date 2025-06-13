<script lang="ts">
  import { redirect } from '@sveltejs/kit';
  import { onMount } from 'svelte';
  import { UserHandler } from '../../../lib/handlers/user';

  onMount(() => {
    const handler = new UserHandler();

    if (handler.IsLoggedIn()) {
      // redirect to login
      redirect(302, '/users/login');
    }
  });
</script>

<svelte:head>
  <title>User</title>
</svelte:head>
