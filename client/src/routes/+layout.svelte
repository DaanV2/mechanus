<script lang="ts">
  import { Navbar, NavBrand, NavHamburger, NavUl, NavLi } from 'flowbite-svelte';
  import '../app.css';

  let { children } = $props();
</script>

{@render children()}
