<script>
  import { Application, Assets, Sprite } from 'pixi.js';
  import { onMount } from 'svelte';

  const app = new Application();

  /**
   *
   * @param from {number}
   * @param to {number}
   * @param delta {number}
   */
  function move(from, to, delta) {
    return from + (to - from) * delta * 0.05;
  }

  /**
   *
   * @param item {Sprite}
   * @param x {number}
   * @param y {number}
   * @param delta {number}
   */
  function moveSprite(item, x, y, delta) {
    item.x = move(item.x, x, delta);
    item.y = move(item.y, y, delta);
  }

  onMount(async () => {
    console.log('mounting');
    await app.init({
      autoDensity: true,
      background: '#1099bb',
      resizeTo: window,
      preference: 'webgpu',
    });

    app.canvas.id = 'render_view';

    document.body.appendChild(app.canvas);
    // Load the bunny texture.
    const arrow_Texture = await Assets.load('/test/arrow.png');

    // Create a new Sprite from an image path.
    const defaultOpt = {
      texture: arrow_Texture,
      scale: 0.2,
      x: app.renderer.width / 2,
      y: app.renderer.height / 2
    };
    const arrowLeftTop = new Sprite({
      anchor: { x: 0, y: 0 },
      rotation: 0,
      ...defaultOpt
    });
    const arrowRightTop = new Sprite({
      anchor: { x: 0, y: 0 },
      rotation: Math.PI / 2,
      ...defaultOpt
    });
    const arrowLeftBottom = new Sprite({
      anchor: { x: 0, y: 0 },
      rotation: Math.PI * 1.5,
      ...defaultOpt
    });
    const arrowRightBottom = new Sprite({
      anchor: { x: 0, y: 0 },
      rotation: Math.PI,
      ...defaultOpt
    });

    // Add to stage.
    app.stage.addChild(arrowLeftTop);
    app.stage.addChild(arrowRightTop);
    app.stage.addChild(arrowLeftBottom);
    app.stage.addChild(arrowRightBottom);

    app.ticker.add((time) => {
      console.log(time.FPS);
      moveSprite(arrowLeftTop, 0, 0, time.deltaTime);
      moveSprite(arrowRightTop, app.renderer.width, 0, time.deltaTime);
      moveSprite(arrowLeftBottom, 0, app.renderer.height, time.deltaTime);
      moveSprite(arrowRightBottom, app.renderer.width, app.renderer.height, time.deltaTime);
    });
  });
</script>

<svelte:window />

<style>
  :global(html, body) {
    height: 100%;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
  }

  :global(#render_view) {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
</style>
