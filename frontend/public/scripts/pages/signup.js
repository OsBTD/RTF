import { apiRequest, PopupMessage } from '../tools.js';
import { Browse } from '../router.js';

export { SignUp }

const SignUp = {
    html: `
    <head>
      <link rel="stylesheet" href="/public/styles/signup.css">
    </head>
    <div id="signup">
      <h1>Sign Up</h1>
      <form id="signupForm">
        <input name="first_name" type="text" placeholder="First Name">
        <span class="error" data-for="first_name"></span><br>

        <input name="last_name" type="text" placeholder="Last Name">
        <span class="error" data-for="last_name"></span><br>

        <input name="username" type="text" placeholder="Username">
        <span class="error" data-for="username"></span><br>

        <input name="email" type="email" placeholder="Email">
        <span class="error" data-for="email"></span><br>

        <input name="birth_date" type="date" placeholder="Birthday">
        <span class="error" data-for="birth_date"></span><br>

        <select name="gender">
          <option value="">Select Gender</option>
          <option value="male">Male</option>
          <option value="female">Female</option>
        </select>
        <span class="error" data-for="gender"></span><br>

        <input name="password" type="password" placeholder="Password">
        <span class="error" data-for="password"></span><br>

        <input name="repeated_password" type="password" placeholder="Confirm Password">
        <span class="error" data-for="repeated_password"></span><br>

        <button type="submit">Sign Up</button>
      </form>
      <p>Already have an account? <a href="/signin" data-link>Sign In</a></p>
    </div>
    `
  ,

  setup: () => {
    const form = document.getElementById('signupForm');
    form.addEventListener('submit', onSubmit);
  }
};

// === HELPERS ===

const onSubmit = async (e) => {
  e.preventDefault();

  const form = e.target;
  const userInfo = readFields(form);
  clearErrors(form);

  const errors = signupValidate(userInfo);
  if (Object.keys(errors).length) {
    showErrors(form, errors);
    return;
  }

  const { status, data, error } = await apiRequest("/signup", userInfo, 'POST');

  if (status === 201 || status === 200) {
    PopupMessage("User created successfully. Please sign in.", 'success');
    Browse("/signin");
  } else {
    handleError({ status, message: error || data });
  }
};


const readFields = (form) => {
  const data = new FormData(form);
  return {
    first_name: data.get('first_name')?.trim(),
    last_name: data.get('last_name')?.trim(),
    username: data.get('username')?.trim(),
    email: data.get('email')?.trim(),
    birth_date: data.get('birth_date'),
    gender: data.get('gender'),
    password: data.get('password'),
    repeated_password: data.get('repeated_password'),
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

const handleError = (err) => {
  switch (err.status) {
    case 400:
      PopupMessage("Invalid input. Please review your form.");
      break;
    case 500:
      PopupMessage("Internal server error. Please try again later.");
      break;
    default:
      PopupMessage("Unexpected error occurred.");
  }
};

const signupValidate = (userInfo) => {
  const errors = {};

  if (!/^[A-Za-z]{3,}$/.test(userInfo.first_name || ''))
    errors.first_name = 'First name must be 3+ letters';

  if (!/^[A-Za-z]{3,}$/.test(userInfo.last_name || ''))
    errors.last_name = 'Last name must be 3+ letters';

  if (!/^[A-Za-z0-9_]{3,20}$/.test(userInfo.username || ''))
    errors.username = 'Username 3-20 chars, letters/numbers/_ only';

  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(userInfo.email || ''))
    errors.email = 'Invalid email';

  if (!/^\d{4}-\d{2}-\d{2}$/.test(userInfo.birth_date || '')) {
    errors.birth_date = 'Invalid date format';
  } else {
    const bd = new Date(userInfo.birth_date);
    const age = new Date().getFullYear() - bd.getFullYear();
    if (isNaN(bd) || age < 13 || (age === 13 && Date.now() < bd.setFullYear(bd.getFullYear() + 13)))
      errors.birth_date = 'You must be at least 13';
  }

  if (!/^(male|female)$/.test(userInfo.gender || ''))
    errors.gender = 'Select male or female';

  if (!/^.*(?=.{8,})(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_\-+=]).*$/.test(userInfo.password || ''))
    errors.password = 'Password 8+ chars, include upper, digit & special';

  if (userInfo.repeated_password !== userInfo.password)
    errors.repeated_password = 'Passwords do not match';

  return errors;
};
