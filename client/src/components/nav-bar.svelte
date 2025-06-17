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
    Button,
    DarkMode
  } from 'flowbite-svelte';
  import { UserHandler } from '../lib/handlers/user';
  import { onMount } from 'svelte';

  let user: UserHandler | undefined = $state(undefined);

  onMount(() => {
    user = new UserHandler();
  });
</script>

<Navbar class="bg-primary shadow-2xl">
  <NavBrand href="/">
    <!-- <img src="/images/flowbite-svelte-icon-logo.svg" class="me-3 h-6 sm:h-9" alt="Flowbite Logo" /> -->
    <span class="text-text-100 self-center text-xl font-semibold whitespace-nowrap">Mechanus</span>
  </NavBrand>

  <div class="flex flex-row items-center">
    <!-- User or Login -->
    {#if user?.hasLoggedinUser}
      <div class="flex items-center md:order-3">
        <Avatar id="avatar-menu" src="/images/profile-picture-3.webp" />
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
      <Button
        class="bg-quaternary hover:bg-secondary hover:text-quaternary text-secondary mx-1 flex transition-colors md:order-3"
        href="/users/login">Login</Button
      >
    {/if}

    <DarkMode
      class="bg-quaternary hover:bg-secondary hover:text-quaternary text-secondary mx-1 transition-colors md:order-2"
    />

    <!-- Navigation -->
    <NavHamburger
      class="bg-quaternary hover:bg-secondary hover:text-quaternary text-secondary order-1"
    />
    <NavUl class="text-text-100 order-1">
      <NavLi
        activeClass="text-secondary hover:bg-tertiary hover:text-quaternary transition-colors"
        nonActiveClass="hover:text-text-100 text-secondary transition-colors"
        href="/">Home</NavLi
      >
      <NavLi
        activeClass="text-secondary hover:bg-tertiary hover:text-quaternary transition-colors"
        nonActiveClass="hover:text-text-100 text-secondary transition-colors"
        href="/devices">Devices</NavLi
      >
      <NavLi
        activeClass="text-secondary hover:bg-tertiary hover:text-quaternary transition-colors"
        nonActiveClass="hover:text-text-100 text-secondary transition-colors"
        href="/about">About</NavLi
      >
    </NavUl>
  </div>
</Navbar>
