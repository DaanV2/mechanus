<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { UserHandler } from '../../../lib/handlers/user';
  import { User } from '../../../proto/users/v1/users_pb';

  let user = $state<User | undefined>(undefined);

  onMount(async () => {
    const handler = new UserHandler();

    if (!handler.hasLoggedinUser) {
      // redirect to login
      return goto('/users/login');
    }

    const data = await handler.serverData();
    user = data.user;
  });
</script>

<svelte:head>
  <title>User</title>
</svelte:head>
