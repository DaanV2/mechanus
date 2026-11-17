<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
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

  onMount(async () => {
    currentUser = await userHandler.restore();
  });

  // Dark text on the orange bar (white only reaches 3.7:1); the active page is bold and underlined
  const navActive =
    'text-primary-950 font-bold underline decoration-2 underline-offset-4 md:hover:bg-transparent hover:bg-primary-950/15 rounded transition-colors';
  const navInactive =
    'text-primary-950 md:hover:text-white md:hover:bg-transparent hover:bg-primary-950/15 rounded transition-colors';
  const navButton =
    'bg-quaternary text-secondary hover:bg-secondary hover:text-quaternary dark:bg-secondary dark:text-quaternary dark:hover:bg-quaternary dark:hover:text-secondary mx-1 transition-colors';

  // Nav links have no trailing slash, the router adds one (`trailingSlash = 'always'`)
  let activeUrl = $derived(page.url.pathname.replace(/(.)\/$/, '$1'));

  async function logout() {
    await userHandler.logout();
    currentUser = UserState.LOGGED_OUT;
    await goto('/users/login');
  }
</script>

<Navbar class="bg-primary dark:bg-primary">
  <NavBrand href="/">
    <!-- <img src="/images/flowbite-svelte-icon-logo.svg" class="me-3 h-6 sm:h-9" alt="Flowbite Logo" /> -->
    <span class="self-center text-xl font-semibold whitespace-nowrap text-white">Mechanus</span>
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
          {#if userData.user.roles?.includes('admin')}
            <DropdownItem href="/users">Users</DropdownItem>
            <DropdownItem href="/devices">Devices</DropdownItem>
          {/if}
        </DropdownGroup>
        <DropdownHeader onclick={logout}>Sign out</DropdownHeader>
      </Dropdown>
    {:else}
      <Button class="{navButton} flex md:order-3" href="/users/login">Login</Button>
    {/if}

    <DarkMode class="{navButton} md:order-2" />

    <!-- Navigation -->
    <NavHamburger class="{navButton} order-1" />
    <NavUl
      {activeUrl}
      class="order-1"
      classes={{
        ul: 'bg-primary md:bg-transparent border-primary-950/20 dark:bg-primary dark:border-primary-950/20'
      }}
    >
      <NavLi activeClass={navActive} nonActiveClass={navInactive} href="/">Home</NavLi>

      {#if userData}
        <NavLi activeClass={navActive} nonActiveClass={navInactive} href="/campaigns"
          >Campaigns</NavLi
        >
      {/if}
      <NavLi activeClass={navActive} nonActiveClass={navInactive} href="/devices">Devices</NavLi>
      <NavLi activeClass={navActive} nonActiveClass={navInactive} href="/about">About</NavLi>
    </NavUl>
  </div>
</Navbar>
