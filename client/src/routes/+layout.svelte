<script lang="ts">
  import '../app.css';

  let { children } = $props();
</script>

<!-- A column filling the viewport, so a page's footer (`mt-auto`) sits at the bottom of short pages -->
<div class="flex min-h-screen flex-col">
  {@render children()}
</div>
