import { expect, test, type APIRequestContext, type Page } from "@playwright/test";
import { callRPC, login as apiLogin } from "../../lib/api/client";
import { User } from "../../lib/users/create";
import { login } from "../../lib/users/login";

// The example scene from docs/examples/battlemap: 16 x 9 cells, so at the default 1280 x 720 viewport
// one grid cell is 80 screen pixels.
const SCENE = "example-tavern";
const DEVICE = "example-tv";
const CELL = 80;

const FLOOR = "#8a6a4a";
const FLOOR_NIGHT = "#1f2a44";
const RUG = "#7c2d2d";
const BARREL = "#3f2a14";
const BARKEEP = "#16a34a";
const HERO = "#2563eb";
const SPY = "#dc2626";

// The tests share the one active scene and its saved state, so they run one after the other and put back what
// they changed
test.describe.configure({ mode: "serial" });

type RGB = [number, number, number];

/** The colours at the given screen positions, read from a screenshot of the canvas as the user sees it. */
async function colorsAt(page: Page, points: [number, number][]): Promise<RGB[]> {
  const png = await page.locator("canvas").screenshot();

  return page.evaluate(
    async ({ data, points }) => {
      const image = new Image();
      image.src = `data:image/png;base64,${data}`;
      await image.decode();

      const canvas = document.createElement("canvas");
      canvas.width = image.width;
      canvas.height = image.height;
      const context = canvas.getContext("2d")!;
      context.drawImage(image, 0, 0);

      return points.map(([x, y]) => [...context.getImageData(x, y, 1, 1).data.slice(0, 3)] as RGB);
    },
    { data: png.toString("base64"), points },
  );
}

function isColor(actual: RGB, hex: string): boolean {
  const expected = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16));

  return actual.every((v, i) => Math.abs(v - expected[i]) <= 6);
}

/** The screen position of the center of a grid cell. */
function cell(x: number, y: number): [number, number] {
  return [(x + 0.5) * CELL, (y + 0.5) * CELL];
}

/**
 * Expects each named point to have its colour. All points are read from one screenshot, as a screenshot
 * takes seconds where the browser renders in software, such as on CI.
 */
async function expectColors(page: Page, expected: Record<string, [[number, number], string]>) {
  const names = Object.keys(expected);

  await expect
    .poll(
      async () => {
        const colors = await colorsAt(page, names.map((name) => expected[name][0]));

        return names.filter((name, i) => !isColor(colors[i], expected[name][1]));
      },
      { message: "points with the wrong colour", timeout: 15000 },
    )
    .toEqual([]);
}

test.describe("battle map on a TV", { tag: ["@scenes", "@battlemap"] }, () => {
  test("shows the layers, grid and entities of the active scene", async ({ page, request }) => {
    test.setTimeout(60000);
    let token = "";

    await test.step("by activating the example scene", async () => {
      token = await apiLogin(request, User.createAdmin());
      await callRPC(request, "scenes.v1.SceneService/Activate", { sceneId: SCENE }, token);
      await callRPC(request, "scenes.v1.SceneService/SetVariant", {}, token);
    });

    await test.step("by opening the screen of the device, which takes a logged in user", async () => {
      await login(page, User.createAdmin());
      await page.goto(`/views/devices/?device=${DEVICE}`);
      await expect(page.locator("canvas")).toBeVisible();
    });

    await test.step("image layers are drawn in order", async () => {
      await expectColors(page, {
        "the floor image": [cell(3, 6), FLOOR],
        "the rug image, on top of the floor": [cell(7, 4), RUG],
      });
    });

    await test.step("grid lines are drawn", async () => {
      const y = cell(3, 6)[1];
      const colors = await colorsAt(page, [316, 317, 318, 319, 320, 321, 322, 323, 324].map((x) => [x, y]));

      expect(colors.some((c) => !isColor(c, FLOOR)), "a grid line between cell 3 and 4").toBe(true);
    });

    await test.step("instance groups draw one asset at every position", async () => {
      await expectColors(page, {
        "the first barrel": [cell(1, 1), BARREL],
        "the second barrel": [cell(2, 1), BARREL],
        "the third barrel": [cell(14, 7), BARREL],
      });
    });

    await test.step("entities are drawn, and hidden ones aren't", async () => {
      const [x, y] = cell(4, 4);

      await expectColors(page, {
        "the barkeep's token": [cell(11, 2), BARKEEP],
        "the hero's placeholder token": [[x - 28, y], HERO],
        "the floor, where the hidden spy stands": [cell(5, 7), FLOOR],
      });
    });

    await test.step("switching variant updates the TV", async () => {
      await callRPC(request, "scenes.v1.SceneService/SetVariant", { variantId: "night" }, token);
      await expectColors(page, {
        "the night floor": [cell(3, 6), FLOOR_NIGHT],
        "the barkeep's token at night": [cell(11, 2), BARKEEP],
      });
    });
  });

  test("does not serve the map file, which names where hidden entities are", async ({ request }) => {
    const response = await request.get(`/assets/scenes/${SCENE}/maps/tavern.map.json`);

    expect(response.status()).toBe(404);
  });
});

test.describe("live play on the battle map", { tag: ["@scenes", "@battlemap"] }, () => {
  let token = "";

  /** Puts the scene back as the other tests expect it: hero at the map's position, spy hidden. */
  async function resetScene(request: APIRequestContext) {
    await callRPC(request, "scenes.v1.SceneService/Activate", { sceneId: SCENE }, token);
    await callRPC(request, "scenes.v1.SceneService/SetVariant", {}, token);
    await callRPC(request, "scenes.v1.SceneService/MoveEntity", { entityId: "hero", position: { x: 4, y: 4 } }, token);
    await callRPC(request, "scenes.v1.SceneService/SetEntityHidden", { entityId: "spy", hidden: true }, token);
  }

  /** Opens the screen of the TV as a logged in user, who can move tokens on it. */
  async function openTV(page: Page) {
    await page.goto(`/views/devices/?device=${DEVICE}`);
    await expect(page.locator("canvas")).toBeVisible();
    await expectColors(page, { "the hero's placeholder token": [heroPoint(4, 4), HERO] });
  }

  /** A point on the hero's placeholder disc, left of the initial drawn in its middle. */
  function heroPoint(x: number, y: number): [number, number] {
    const [cx, cy] = cell(x, y);

    return [cx - 28, cy];
  }

  test.beforeEach(async ({ request }) => {
    token = await apiLogin(request, User.createAdmin());
    await resetScene(request);
  });

  test.afterEach(async ({ request }) => {
    await resetScene(request);
  });

  test("dragging a token moves it on every screen and stores it in the scene state", async ({ page, request }) => {
    test.setTimeout(90000);
    await login(page, User.createAdmin());
    await openTV(page);
    const other = await page.context().newPage();
    await openTV(other);

    await test.step("by dragging the hero two cells right and down", async () => {
      const [fromX, fromY] = cell(4, 4);
      const [toX, toY] = cell(6, 6);
      await page.mouse.move(fromX, fromY);
      await page.mouse.down();
      await page.mouse.move(toX, toY, { steps: 10 });
      await page.mouse.up();
    });

    await test.step("the token moved on the screen it was dragged on", async () => {
      await expectColors(page, {
        "the hero at its new cell": [heroPoint(6, 6), HERO],
        "the floor where it stood": [heroPoint(4, 4), FLOOR],
      });
    });

    await test.step("and on the other screen showing the TV", async () => {
      await expectColors(other, {
        "the hero at its new cell": [heroPoint(6, 6), HERO],
        "the floor where it stood": [heroPoint(4, 4), FLOOR],
      });
    });

    await test.step("the position is kept in the scene state", async () => {
      const resp = await callRPC<{ active: { entities: { id: string; position?: { x?: number; y?: number } }[] } }>(
        request,
        "scenes.v1.SceneService/GetActive",
        {},
        token,
      );
      const hero = resp.active.entities.find((e) => e.id === "hero");
      expect(hero?.position).toEqual({ x: 6, y: 6 });
    });

    await test.step("a screen opened later draws it at its new position", async () => {
      const later = await page.context().newPage();
      await later.goto(`/views/devices/?device=${DEVICE}`);
      await expectColors(later, { "the hero at its new cell": [heroPoint(6, 6), HERO] });
    });
  });

  test("the GM reveals and hides an entity from the GM view", async ({ page }) => {
    test.setTimeout(90000);
    await login(page, User.createAdmin());
    await openTV(page);
    const gm = await page.context().newPage();
    await gm.goto("/views/game-master/");
    const spy = gm.getByTestId("entity-spy");
    await expect(spy).toContainText("Hidden");

    await test.step("the TV doesn't show the hidden spy", async () => {
      await expectColors(page, { "the floor, where the hidden spy stands": [cell(5, 7), FLOOR] });
    });

    await test.step("revealing the spy shows it on the TV", async () => {
      await spy.getByRole("button", { name: "Reveal" }).click();
      await expect(spy).toContainText("Visible");
      await expectColors(page, { "the spy's token": [cell(5, 7), SPY] });
    });

    await test.step("hiding it again removes it from the TV", async () => {
      await spy.getByRole("button", { name: "Hide" }).click();
      await expect(spy).toContainText("Hidden");
      await expectColors(page, { "the floor, where the hidden spy stands": [cell(5, 7), FLOOR] });
    });
  });
});
