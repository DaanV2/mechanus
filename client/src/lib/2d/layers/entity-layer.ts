import { Container, Graphics, Particle, ParticleContainer, Sprite, Text } from 'pixi.js';
import {
  EntityKind,
  type Entity,
  type InstanceGroup,
  type Layer
} from '../../../proto/screens/v1/state_pb';
import { degreesToRadians, type MapSpace } from './layer';
import type { TextureCache } from './textures';

const PLACEHOLDER_COLORS: Partial<Record<EntityKind, string>> = {
  [EntityKind.PLAYER]: '#2563eb',
  [EntityKind.NPC]: '#16a34a',
  [EntityKind.ENEMY]: '#dc2626',
  [EntityKind.OBJECT]: '#737373'
};

/** Draws an entity layer in map pixels: its instance groups, with the entity tokens on top of them. */
export async function createEntityLayer(
  layer: Layer,
  space: MapSpace,
  textures: TextureCache
): Promise<Container> {
  const root = new Container({ label: layer.id });

  const groups = await Promise.all(
    layer.instanceGroup.map((group) => createInstanceGroup(group, space, textures))
  );
  const tokens = await Promise.all(layer.entity.map((e) => createToken(e, space, textures)));
  root.addChild(...groups.filter((g) => g !== undefined), ...tokens);

  return root;
}

/**
 * All instances of a group share one texture, so a particle container draws them together: this is
 * what allows a map to hold thousands of them. They don't move, so nothing is updated per frame.
 */
async function createInstanceGroup(
  group: InstanceGroup,
  space: MapSpace,
  textures: TextureCache
): Promise<ParticleContainer | undefined> {
  const texture = await textures.load(group.asset);
  if (!texture) return undefined;

  const width = (group.size?.x ?? 1) * space.cell;
  const height = (group.size?.y ?? 1) * space.cell;

  const container = new ParticleContainer({
    label: group.id,
    texture,
    dynamicProperties: {
      position: false,
      rotation: false,
      vertex: false,
      uvs: false,
      color: false
    },
    particles: group.instance.map(
      (i) =>
        new Particle({
          texture,
          x: space.x(i.x) + width / 2,
          y: space.y(i.y) + height / 2,
          anchorX: 0.5,
          anchorY: 0.5,
          scaleX: width / texture.width,
          scaleY: height / texture.height,
          rotation: degreesToRadians(i.rotation)
        })
    )
  });
  // Static properties are only uploaded after an update, passing particles to the constructor does not trigger one
  container.update();

  return container;
}

/** Draws a token centred on its cells, so it rotates around its center. */
async function createToken(
  entity: Entity,
  space: MapSpace,
  textures: TextureCache
): Promise<Container> {
  const width = (entity.size?.x ?? 1) * space.cell;
  const height = (entity.size?.y ?? 1) * space.cell;

  const texture = await textures.load(entity.asset);
  let token: Container;
  if (texture) {
    const sprite = new Sprite(texture);
    sprite.anchor.set(0.5);
    sprite.setSize(width, height);
    token = sprite;
  } else {
    token = createPlaceholder(entity, width, height);
  }

  token.label = entity.id;
  token.position.set(
    space.x(entity.position?.x ?? 0) + width / 2,
    space.y(entity.position?.y ?? 0) + height / 2
  );
  token.rotation = degreesToRadians(entity.rotation);

  return token;
}

/** A coloured disc with the entity's initial, for entities without a token image. */
function createPlaceholder(entity: Entity, width: number, height: number): Container {
  const root = new Container();

  const disc = new Graphics()
    .ellipse(0, 0, width * 0.45, height * 0.45)
    .fill(PLACEHOLDER_COLORS[entity.kind] ?? '#737373')
    .stroke({ width: Math.max(1, width * 0.04), color: '#ffffff' });
  const initial = new Text({
    text: entity.name.charAt(0).toUpperCase(),
    style: { fontSize: height * 0.5, fontWeight: 'bold', fill: '#ffffff' }
  });
  initial.anchor.set(0.5);
  root.addChild(disc, initial);

  return root;
}
