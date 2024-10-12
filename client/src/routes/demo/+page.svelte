<script lang="ts">
  // Define an array of subpage objects
  const subpages = [{ name: 'Size check', path: 'demo/size' }];
</script>

<div class="centered-container">
  <div class="box-container">
    <ul class="subpage-list">
      {#each subpages as { name, path }}
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

  .subpage-link {
    text-decoration: none;
    color: #007bff;
    font-size: 1rem;
    transition: color 0.3s;
  }

  .subpage-link:hover {
    color: #0056b3;
  }
</style>
