import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { licenses } from './scripts/vite-plugin-licenses';

export default defineConfig({
  plugins: [tailwindcss(), sveltekit(), licenses()]
});
