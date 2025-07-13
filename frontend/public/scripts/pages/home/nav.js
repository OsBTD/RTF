import { apiRequest, PopupMessage } from "../../tools.js";
import { Browse } from "../../router.js";

export { NavBar, ThemeToggle };

const NavBar = {
  html: `
  <head>
    <link rel="stylesheet" href="/public/styles/nav.css">
  </head>
  <nav>
    <div id="left">
      <span class="material-symbols-outlined" alt="Menu Icon">segment</span>
    </div>
    <div id="center">
      <a href="/" data-link>
        <img src="/public/images/echohub-logo.png" alt="echohub community" width="300">
      </a>
    </div>
    <div id="right">
      <div id="profile">
      <span id="uname"></span>
        <span id="profile-icon" class="material-symbols-outlined" alt="Profile">account_circle</span>
        <div class="dropdown" id="dropdownMenu">
          <div id="user-info">
            <img id="uimg" alt="User Image">
            <div>
              <span id="unamesub"></span>
              <span id="uemail"></span>
            </div>
          </div>
          <a id="new-post" href="/newpost" data-link>
            <span class="material-symbols-outlined" alt="New Post">add</span>
            <span>New Post</span>
          </a>
          <button id="dark-toggle">
            <input type="checkbox" id="checkboxInput">
            <label for="checkboxInput" class="toggleSwitch"></label>
            <span>Dark Mode</span>
          </button>
          <button id="logout">
            <span class="material-symbols-outlined" alt="Logout">logout</span>
            <span>Logout</span>
          </button>
        </div>
      </div>
    </div>
  </nav>`,

  setup: () => {
    SidebarToggle();
    ProfileDropdown();
    UserInfo();
    ThemeToggle();
    Logout();
  }
};

const SidebarToggle = () => {
  const toggleBtn = document.getElementById('left');
  const sidebar = document.getElementById('categories-container');
  toggleBtn.addEventListener('click', () => sidebar?.classList.toggle('show'));
};

const ProfileDropdown = () => {
  const profileBtn = document.getElementById('profile');
  const dropdownMenu = document.getElementById('dropdownMenu');

  profileBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    dropdownMenu.classList.toggle('show');
  });

  document.addEventListener('click', (e) => {
    if (!profileBtn.contains(e.target) && !dropdownMenu.contains(e.target)) {
      dropdownMenu.classList.remove('show');
    }
  });
};

const UserInfo = () => {
  const userImg = document.getElementById('uimg');
  const userNameSpan = document.getElementById('uname');
  const userNameSubSpan = document.getElementById('unamesub');
  const userEmailSpan = document.getElementById('uemail');

  const username = localStorage.getItem('username');
  const email = localStorage.getItem('email');
  const profileImg = localStorage.getItem('profile_img');

  if (profileImg) userImg.src = profileImg;
  if (username) userNameSpan.textContent= userNameSubSpan.textContent = username;
  if (email) userEmailSpan.textContent = email;
};

const ThemeToggle = () => {
  const darkToggle = document.getElementById('checkboxInput');

  const applyTheme = (theme) => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
   if (darkToggle) darkToggle.checked = theme === 'dark';
  };

  const savedTheme = localStorage.getItem('theme');
  if (savedTheme) {
    applyTheme(savedTheme);
  } else {
    const prefersDark = !window.matchMedia('(prefers-color-scheme: dark)').matches;
    applyTheme(prefersDark ? 'dark' : 'light');
  }

  if (darkToggle) darkToggle.addEventListener('change', () => {
    applyTheme(darkToggle.checked ? 'dark' : 'light');
  });
};

const Logout = () => {
  const logoutBtn = document.getElementById('logout');

  logoutBtn.addEventListener('click', async () => {
    const { status, error } = await apiRequest('/signout', {}, 'DELETE');

    if (status === 200 || status === 204) {
      localStorage.clear();
      Browse("/signin");
      PopupMessage("Sign out successful", 'success');
    } else if (status === 401) {
      localStorage.clear();
      Browse("/signin");
      PopupMessage("Session expired. Signed out.", 'info', 15);
    } else {
      PopupMessage("Failed to sign out. Please try again.");
    }
  });
};
