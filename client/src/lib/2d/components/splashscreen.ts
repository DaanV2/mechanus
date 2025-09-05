import { Container, Graphics, Rectangle, Text, TextStyle } from 'pixi.js';

export class SplashScreen extends Container {
  public readonly background: Graphics;
  public readonly title: Text;
  public readonly titleStyle: TextStyle;
  public readonly subtitleStyle: TextStyle;
  public readonly subtitle: Text;

  constructor() {
    super();

    this.background = new Graphics().rect(0, 0, 800, 400).fill('#000000');

    this.titleStyle = new TextStyle({
      fontSize: 96,
      fill: '#FFFFFF',
      fontWeight: 'bold'
    });
    this.subtitleStyle = new TextStyle({
      fontSize: 32,
      fill: '#FFFFFF',
      fontWeight: 'bold'
    });
    this.title = new Text({
      text: 'Machnus',
      style: this.titleStyle
    });
    this.subtitle = new Text({
      text: '',
      style: this.subtitleStyle
    });

    this.addChild(this.background, this.title, this.subtitle);
  }

  public handleResize(screen: Pick<Rectangle, 'width' | 'height'>) {
    const x = screen.width / 2;
    const y = screen.height / 2;

    this.background.setSize(screen.width, screen.height);
    this.title.position = { x: x, y: y - (this.title.height + 2) };
    this.subtitle.position = {
      x: x,
      y: y + (this.subtitle.height + 2)
    };
  }
}
