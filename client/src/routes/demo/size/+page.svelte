<script lang="ts">
  import { Application, Assets, Sprite } from 'pixi.js';
  import { onMount } from 'svelte';

  /**
   * Size check: four arrows move into the corners of the canvas. When the canvas matches the
   * viewport, every arrow tip sits exactly in a corner of the screen and the readout shows the same
   * size twice.
   *
   * The canvas is rendered into a fixed, full-viewport container and sized to that container (not
   * to `window`): `window.innerWidth` includes a vertical scrollbar and the canvas used to be
   * appended below the page in <body>, which made it too wide and too high.
   */
  let container: HTMLDivElement;
  let canvas: HTMLCanvasElement;
  let canvasSize = $state('');
  let viewportSize = $state('');

  function move(from: number, to: number, delta: number) {
    return from + (to - from) * delta * 0.05;
  }

  function moveSprite(item: Sprite, x: number, y: number, delta: number) {
    item.x = move(item.x, x, delta);
    item.y = move(item.y, y, delta);
  }

  onMount(() => {
    const app = new Application();
    let destroyed = false;

    async function start() {
      await app.init({
        canvas,
        autoDensity: true,
        resolution: window.devicePixelRatio,
        background: '#1099bb',
        resizeTo: container,
        preference: 'webgpu'
      });
      if (destroyed) {
        app.destroy();

        return;
      }

      const arrowTexture = await Assets.load('/test/arrow.png');
      const defaultOpt = {
        texture: arrowTexture,
        scale: 0.2,
        anchor: { x: 0, y: 0 },
        x: app.screen.width / 2,
        y: app.screen.height / 2
      };
      const arrowLeftTop = new Sprite({ rotation: 0, ...defaultOpt });
      const arrowRightTop = new Sprite({ rotation: Math.PI / 2, ...defaultOpt });
      const arrowLeftBottom = new Sprite({ rotation: Math.PI * 1.5, ...defaultOpt });
      const arrowRightBottom = new Sprite({ rotation: Math.PI, ...defaultOpt });
      app.stage.addChild(arrowLeftTop, arrowRightTop, arrowLeftBottom, arrowRightBottom);

      app.ticker.add((time) => {
        // `screen` is in CSS pixels, `renderer.width` is in device pixels
        const { width, height } = app.screen;
        moveSprite(arrowLeftTop, 0, 0, time.deltaTime);
        moveSprite(arrowRightTop, width, 0, time.deltaTime);
        moveSprite(arrowLeftBottom, 0, height, time.deltaTime);
        moveSprite(arrowRightBottom, width, height, time.deltaTime);

        canvasSize = `${Math.round(width)} × ${Math.round(height)}`;
        viewportSize = `${document.documentElement.clientWidth} × ${document.documentElement.clientHeight}`;
      });
    }

    start();

    return () => {
      destroyed = true;
      // Only destroy once init has finished, `start` handles the other case
      if (app.renderer) app.destroy();
    };
  });
</script>

<svelte:head>
  <title>Mechanus - Size check</title>
</svelte:head>

<div class="fixed inset-0 overflow-hidden" bind:this={container}>
  <canvas class="block" bind:this={canvas}></canvas>
</div>

<div
  class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded bg-black/70 px-4 py-2 text-center font-mono text-sm text-white"
  data-testid="size-readout"
>
  <div>canvas {canvasSize}</div>
  <div>viewport {viewportSize}</div>
</div>
