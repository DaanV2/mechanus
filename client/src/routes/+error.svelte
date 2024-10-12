<script>
  import { page } from '$app/stores';
</script>

<h1>{$page.error?.message ?? "UNKNOWN"}</h1>
