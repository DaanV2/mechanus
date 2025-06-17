# Web Server

this is the client side of the Mechanus project, which is the UI for the players and DM in a tabletop RPG game. It connects to the server to display information, manage player interactions, and provide a seamless gaming experience.

The webpages are rendered as static files, which are served by the server. The client is built using Svelte, Tailwind CSS and connectRPC for gRPC communication to the server. We use Playwright for end-to-end testing to ensure the client behaves correctly across different browsers and devices. and [Flowbite](https://flowbite-svelte.com/) for UI components.

## Development

To set up the client for development, you can use the following commands:

```bash
npm install
npm run dev

# To start up the server: (this requires golang)
cd ../server
make start-server
```
