<script lang="ts">
  import { userHandler, UserState } from '$lib/handlers/user';
  import {
    Avatar,
    Button,
    DarkMode,
    Dropdown,
    DropdownGroup,
    DropdownHeader,
    DropdownItem,
    Navbar,
    NavBrand,
    NavHamburger,
    NavLi,
    NavUl
  } from 'flowbite-svelte';
  import { onMount } from 'svelte';

  let currentUser: UserState = $state(UserState.LOGGED_OUT);
  let userData = $derived(currentUser.loggedin ? currentUser.data : undefined);

  onMount(() => {
    currentUser = userHandler.current;
  });

  async function logout() {
    return userHandler.logout();
  }
</script>

<Navbar class="bg-primary">
  <NavBrand href="/">
    <!-- <img src="/images/flowbite-svelte-icon-logo.svg" class="me-3 h-6 sm:h-9" alt="Flowbite Logo" /> -->
    <span class="text-text-100 self-center text-xl font-semibold whitespace-nowrap">Mechanus</span>
  </NavBrand>

  <div class="flex flex-row items-center">
    <!-- User or Login -->
    {#if userData}
      <div class="flex items-center md:order-3">
        <Avatar id="avatar-menu">{userData.user.name.slice(0, 2)}</Avatar>
      </div>
      <Dropdown placement="bottom" triggeredBy="#avatar-menu">
        <DropdownHeader>
          <span class="block text-sm">{userData.user.name}</span>
        </DropdownHeader>
        <DropdownGroup>
          <DropdownItem href="/users/login">Profile</DropdownItem>
        </DropdownGroup>
        <DropdownHeader onclick={logout}>Sign out</DropdownHeader>
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

      {#if userData}
        <NavLi
          activeClass="text-secondary hover:bg-tertiary hover:text-quaternary transition-colors"
          nonActiveClass="hover:text-text-100 text-secondary transition-colors"
          href="/campaigns">Campaigns</NavLi
        >
      {/if}
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
