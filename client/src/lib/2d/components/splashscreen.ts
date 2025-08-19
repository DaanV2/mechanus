import { Container, Graphics, Text, TextStyle, type Application } from 'pixi.js';

interface SplashScreenOptions {
  title: string;
  subtitle: string;
  background: string;
  textColor: string;
}

const defOpts = {
  title: 'Mechanus',
  subtitle: 'Virtual table top software',
  textColor: '#ffffff',
  background: '#000000'
};

export function createSplashScreen(
  app: Application,
  opts?: Partial<SplashScreenOptions>
): Container {
  opts = {
    ...defOpts,
    ...opts
  };

  const container = new Container();

  if ('background' in opts) {
    container.addChild(
      new Graphics().rect(0, 0, app.screen.width, app.screen.height).fill(opts['background'])
    );
  }

  if ('title' in opts) {
    container.addChild(
      new Text({
        text: opts['title'],
        position: { x: app.screen.width / 2, y: app.screen.height / 2 - 50 },
        style: new TextStyle({
          fontSize: 96,
          fill: opts['textColor'],
          fontWeight: 'bold'
        })
      })
    );
  }

  if ('subtitle' in opts) {
    container.addChild(
      new Text({
        text: opts['subtitle'],
        position: { x: app.screen.width / 2, y: app.screen.height / 2 + 50 },
        style: new TextStyle({
          fontSize: 32,
          fill: opts['textColor'],
          fontWeight: 'bold'
        })
      })
    );
  }

  return container;
}
