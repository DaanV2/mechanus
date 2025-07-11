<script>
  // No props needed, just slot for children
</script>

<div class="w-4/5 justify-center align-middle">
  <!-- svelte-ignore element_invalid_self_closing_tag -->
  <div class="h-3" />

  <div class="fluid-container">
    <slot />
  </div>

  <!-- svelte-ignore element_invalid_self_closing_tag -->
  <div class="h-3" />
</div>

<style>
  .fluid-container {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    align-items: center;
    width: auto;
  }
  .fluid-container > * {
    flex: 0 0 auto;
    min-width: 0;
  }
</style>
