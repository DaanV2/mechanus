<script lang="ts">
  import { Button } from 'flowbite-svelte';
  import type { ComponentProps } from 'svelte';

  /**
   * The app's standard button: dark by default, grey when disabled, and the hover colour says what
   * it does: green for creating / saving, red for deleting / revoking, darker for anything else.
   */
  type Intent = 'create' | 'delete' | 'neutral';

  let {
    intent = 'neutral',
    class: className = '',
    children,
    ...rest
  }: ComponentProps<typeof Button> & { intent?: Intent } = $props();

  const base =
    'text-white bg-primary-800 dark:bg-primary-600 transition-colors focus-within:ring-primary-400 ' +
    'disabled:bg-gray-300 disabled:text-gray-500 disabled:opacity-100 ' +
    'dark:disabled:bg-gray-800 dark:disabled:text-gray-500';

  const hovers: Record<Intent, string> = {
    create: 'enabled:hover:bg-ok-700 dark:enabled:hover:bg-ok-700 [&[href]]:hover:bg-ok-700',
    delete:
      'enabled:hover:bg-error-700 dark:enabled:hover:bg-error-700 [&[href]]:hover:bg-error-700',
    neutral:
      'enabled:hover:bg-primary-950 dark:enabled:hover:bg-primary-500 [&[href]]:hover:bg-primary-950'
  };
</script>

<Button {...rest} class={[base, hovers[intent], className].join(' ')}>
  {@render children?.()}
</Button>
