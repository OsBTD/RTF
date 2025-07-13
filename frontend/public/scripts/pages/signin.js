import { apiRequest, PopupMessage } from '../tools.js';
import { Browse } from '../router.js';

export { SignIn }

let UserInfo;

const SignIn = {
  
    html: `
    <head>
      <link rel="stylesheet" href="/public/styles/signin.css">
    </head>
    <div id="signin">
      <h1>Sign In</h1>
      <form id="signinForm">
        <input name="identifier" type="text" placeholder="Username or Email" />
        <span class="error" data-for="identifier"></span>

        <input name="password" type="password" placeholder="Password" />
        <span class="error" data-for="password"></span>

        <button type="submit">Login</button>
      </form>
      <p>Don't have an account? <a href="/signup" data-link>SignUp</a></p>
    </div>
    `
  ,

  setup: () => {
    const form = document.getElementById('signinForm');
    form.addEventListener('submit', onSubmit);
  }
};

// === HELPERS ===

const onSubmit = async (e) => {
  e.preventDefault();

  const form = e.target;
  const { identifier, password } = readFields(form);
  clearErrors(form);

  const errors = signinValidate({ identifier, password });
  if (Object.keys(errors).length) {
    showErrors(form, errors);
    return;
  }

  const isEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(identifier);
  const payload = {
    username: isEmail ? "" : identifier,
    email: isEmail ? identifier : "",
    password,
  };

  const { status, data, error } = await apiRequest("/signin", payload, 'POST');

  if (status === 200) {
    saveUser(data);
    PopupMessage("Success", 'success');
    Browse("/");
  } else {
    handleError({ status, message: error || data });
  }
};


const readFields = (form) => {
  const data = new FormData(form);
  return {
    identifier: data.get('identifier')?.trim(),
    password: data.get('password')
  };
};

const clearErrors = (form) => {
  form.querySelectorAll('.error').forEach(span => span.textContent = '');
};

const showErrors = (form, errors) => {
  for (const key in errors) {
    const span = form.querySelector(`.error[data-for="${key}"]`);
    if (span) span.textContent = errors[key];
  }
};

const saveUser = (user) => {
  localStorage.setItem('username', user.username);
  localStorage.setItem('email', user.email);
  localStorage.setItem('profile_img', user.profile_img);
  localStorage.setItem('token', user.token);
};

const signinValidate = ({ identifier, password }) => {
  const errors = {};

  if (!identifier) {
    errors.identifier = 'Email or username is required';
  } else {
    const isEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(identifier || '');
    const isUsername = /^[A-Za-z0-9_]{3,20}$/.test(identifier || '');

    if (!isEmail && !isUsername) {
      errors.identifier = 'Enter a valid email or username';
    }
  }

  if (!/^.*(?=.{8,})(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_\-+=]).*$/.test(password || '')) {
    errors.password = 'Password 8+ chars, include upper, digit & special';
  }

  return errors;
};


const handleError = (err) => {
  switch (err.status) {
    case 400:
      PopupMessage("Invalid form input. Please check your fields.");
      break;
    case 401:
      PopupMessage("Incorrect email or password.");
      break;
    case 500:
      PopupMessage("Internal server error. Please try again later.");
      break;
    default:
      PopupMessage("Unexpected error occurred.");
  }
};
