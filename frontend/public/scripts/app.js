import { Browse, RenderRoute } from './router.js';

// Handle link clicks
document.addEventListener('click', (e) => {
  if (e.target.matches('[data-link]')) {
    e.preventDefault();
    const path = e.target.getAttribute('href');
    Browse(path);
  }
});

// popstate >> back/forward
window.addMultiEventListener(['DOMContentLoaded', 'popstate'], () => RenderRoute())

