<script>
  import { page } from '$app/stores';
</script>

<div class="flex min-h-screen grow flex-col items-center justify-center gap-4 px-4 text-center">
  <h1 class="text-ink text-3xl font-bold">{$page.error?.message ?? 'UNKNOWN'}</h1>
  <a href="/" class="link">Back to home</a>
</div>
