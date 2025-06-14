<script lang="ts">
  // Define an array of subpage objects
  const subpages = [{ name: 'Size check', path: '/demo/size' }];
</script>

<div class="centered-container">
  <div class="box-container">
    <ul class="subpage-list">
      {#each subpages as { name, path } (name)}
        <li class="subpage-item">
          <a href={path} class="action-button">{name}</a>
        </li>
      {/each}
    </ul>
  </div>
</div>

<style>
  .subpage-list {
    list-style-type: none;
    padding: 0;
  }

  .subpage-item {
    margin: 0.5rem 0;
  }
</style>
