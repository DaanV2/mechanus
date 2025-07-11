<script lang="ts">
  import { Code } from '@connectrpc/connect';
  import type { MechanusError } from './errors';

  export let error: MechanusError = null;

  $: isVisible = !!error;

  // gRPC error code translation using the Code enum
  const grpcErrorCodeText: Record<Code, string> = {
    [Code.Canceled]: 'Canceled, usually by the user',
    [Code.Unknown]: 'Unknown error',
    [Code.InvalidArgument]: 'Argument invalid regardless of system state',
    [Code.DeadlineExceeded]: 'Operation expired, may or may not have completed',
    [Code.NotFound]: 'Entity not found',
    [Code.AlreadyExists]: 'Entity already exists',
    [Code.PermissionDenied]: 'Operation not authorized',
    [Code.ResourceExhausted]: 'Quota exhausted',
    [Code.FailedPrecondition]: 'Argument invalid in current system state',
    [Code.Aborted]: 'Operation aborted',
    [Code.OutOfRange]: 'Out of bounds',
    [Code.Unimplemented]: 'Operation not implemented or disabled',
    [Code.Internal]: 'Internal error, reserved for serious errors',
    [Code.Unavailable]: 'Unavailable, client should back off and retry',
    [Code.DataLoss]: 'Unrecoverable data loss or corruption',
    [Code.Unauthenticated]: "Request isn't authenticated"
  };

  function codeToText(code: Code) {
    return grpcErrorCodeText[code] || String(code);
  }
</script>

{#if isVisible}
  <div class="error-message">
    {#if error && error instanceof Object && 'code' in error}
      <strong>Login failed:</strong>
      {error.message}
      <ul>
        <li style="text-align:left;">
          <strong>Code:</strong>
          {error.code} ({codeToText(error.code)})
        </li>
        <li style="text-align:left;"><strong>Name:</strong> {error.name}</li>
        {#if 'rawMessage' in error && error.rawMessage}
          <li style="text-align:left;"><strong>Details:</strong> {error.rawMessage}</li>
        {/if}
        {#if 'metadata' in error && error.metadata && Object.keys(error.metadata).length > 0}
          <li style="text-align:left;">
            <strong>Metadata:</strong>
            {JSON.stringify(error.metadata)}
          </li>
        {/if}
        {#if 'cause' in error && error.cause}
          <li style="text-align:left;">
            <strong>Cause:</strong>
            {typeof error.cause === 'string' ? error.cause : JSON.stringify(error.cause)}
          </li>
        {/if}
      </ul>
    {:else}
      <strong>Error:</strong> {typeof error === 'object' ? (error?.message ?? error) : error}
    {/if}
  </div>
{/if}

<style>
  .error-message {
    background: #8b1c1c;
    color: #fff;
    padding: 1em;
    border-radius: 0.5em;
    margin: 1em 0;
    font-size: 1em;
  }
  .error-message ul {
    padding-left: 1.5em;
    margin: 0.5em 0 0 0;
  }
  .error-message li {
    text-align: left;
    margin-bottom: 0.25em;
  }
</style>
