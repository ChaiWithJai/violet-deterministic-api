import './lib/design/tokens.css';
import './lib/design/reset.css';
import './lib/design/typography.css';
import { mount } from 'svelte';
import App from './App.svelte';

const app = mount(App, {
  target: document.getElementById('app')!,
});

export default app;
