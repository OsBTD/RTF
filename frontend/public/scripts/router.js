import { ErrorP } from './pages/error.js';
import { Home } from './pages/home/home.js';
import { SignUp } from './pages/signup.js';
import { SignIn } from './pages/signin.js';
import { NewPost } from './pages/newpost.js';
import { apiRequest } from './tools.js';
import { ThemeToggle } from './pages/home/nav.js';
export { Browse, RenderRoute }

const routes = {
  '/': Home,
  '/signup': SignUp,
  '/signin': SignIn,
  '/newpost': NewPost,
};
 // TODO user must logout from all pages  which means add navbar to all pages
const Browse = (path) => {
  history.pushState({}, '', path);
  RenderRoute(path);
}

const RenderRoute = (path = window.location.pathname) => {
  const view = routes[path];
  const app = document.getElementById('app');
ThemeToggle()

  if (path !== '/signin' && path !== '/signup') {
    const token = localStorage.getItem('token');
    if (!token) {
      apiRequest('/signout', {}, 'DELETE')
        Browse("/signin")
      return;
    }
  }

  if (view) {
    app.innerHTML = view.html;
    view.setup();
  } else {
    app.innerHTML = ErrorP.html;
    ErrorP.setup(404, 'Page Not Found');
  }
};


