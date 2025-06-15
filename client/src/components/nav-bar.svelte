<script lang="ts">
  import {
    Navbar,
    NavBrand,
    Avatar,
    NavHamburger,
    Dropdown,
    DropdownHeader,
    DropdownGroup,
    DropdownItem,
    NavUl,
    NavLi,
    Button
  } from 'flowbite-svelte';
  import { UserHandler } from '../lib/handlers/user';
  import { onMount } from 'svelte';

  let user: UserHandler | undefined = $state(undefined);

  onMount(() => {
    user = new UserHandler();
  });
</script>

<Navbar>
  <NavBrand href="/">
    <!-- <img src="/images/flowbite-svelte-icon-logo.svg" class="me-3 h-6 sm:h-9" alt="Flowbite Logo" /> -->
    <span class="self-center whitespace-nowrap text-xl font-semibold dark:text-white">Mechanus</span
    >
  </NavBrand>

  {#if user?.hasLoggedinUser}
    <div class="flex items-center md:order-2">
      <Avatar id="avatar-menu" src="/images/profile-picture-3.webp" />
      <NavHamburger />
    </div>
    <Dropdown placement="bottom" triggeredBy="#avatar-menu">
      <DropdownHeader>
        <span class="block text-sm">Bonnie Green</span>
        <span class="block truncate text-sm font-medium">name@flowbite.com</span>
      </DropdownHeader>
      <DropdownGroup>
        <DropdownItem>Dashboard</DropdownItem>
        <DropdownItem>Settings</DropdownItem>
        <DropdownItem>Earnings</DropdownItem>
      </DropdownGroup>
      <DropdownHeader>Sign out</DropdownHeader>
    </Dropdown>
  {:else}
    <Button class="flex items-center md:order-2" href="/users/login">Login</Button>
  {/if}

  <NavUl>
    <NavLi href="/">Home</NavLi>
    <NavLi href="/devices">Devices</NavLi>
    <NavLi href="/about">About</NavLi>
  </NavUl>
</Navbar>
