import { apiRequest, PopupMessage } from '../tools.js';
import { Browse } from '../router.js';
import { NavBar } from './home/nav.js';

export { NewPost }

const maxCategories = 3;
const selected = new Set();

const NewPost = {
  
    html:  NavBar.html + `
      <head>
        <link rel="stylesheet" href="/public/styles/newpost.css">
      </head>
      <div id="newpost">
        <h1>New Post</h1>
        <form id="newpost-form">
          <input id="title" name="title" type="text" placeholder="Title.." />
          <span class="error" data-for="title"></span>

          <textarea id="content" name="content" placeholder="Type something..."></textarea>
          <span class="error" data-for="content"></span>

          <div id="categories-selection-container"></div>
          <span class="error" data-for="categories"></span>

          <button type="submit">Submit</button>
        </form>
      </div>
    `
  ,

  setup: async () => {
    NavBar.setup()
    const form = document.getElementById('newpost-form');
    const categoriesContainer = document.getElementById('categories-selection-container');

    
    try {
       const response = await apiRequest('/categories', { target: 'all' }, 'POST');
       renderCategoryOptions(categoriesContainer, response.data);
    } catch (err) {
      PopupMessage("Failed to load categories.");
      return;
    }

    form.addEventListener('submit', onSubmit);
  }
};

// === HELPERS ===

const renderCategoryOptions = (container, categories) => {
  console.log(categories);
  
  categories.forEach(cat => {
    const label = document.createElement('label');
    label.classList.add('category-option');

    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.name = 'categories';
    checkbox.value = cat.id;

    checkbox.addEventListener('change', () => {
      const id = Number(checkbox.value);
      if (checkbox.checked) {
        if (selected.size >= maxCategories) {
          checkbox.checked = false;
          PopupMessage(`Select max ${maxCategories} categories.`);
        } else {
          selected.add(id);
        }
      } else {
        selected.delete(id);
      }
    });

    const span = document.createElement('span');
    span.textContent = cat.name;

    const icon = document.createElement('span');
    icon.classList.add('material-symbols-outlined');
    icon.textContent = cat.icon;
    icon.setAttribute('aria-label', cat.name);

    label.appendChild(icon);
    label.appendChild(span);
    label.appendChild(checkbox);
    container.appendChild(label);
  });
};

const onSubmit = async (e) => {
  e.preventDefault();
  const form = e.target;

  clearErrors(form);
  const { title, content, categories } = readFields(form);
  const errors = validatePost({ title, content, categories });

  if (Object.keys(errors).length) {
    showErrors(form, errors);
    return;
  }

  const payload = { title, content, categories };

  const { status, data, error } = await apiRequest('/newpost', payload, 'POST');

  if (status === 201 || status === 200) {
    PopupMessage("Post created successfully!", 'success');
    Browse('/');
  } else {
    handleError({ status, message: error || data });
  }
};


// === UTILITIES ===

const readFields = (form) => {
  const title = document.getElementById('title').value.trim();
  const content = document.getElementById('content').value.trim();
  const categories = Array.from(selected).map(id => ({ id }));
  return { title, content, categories };
};

const validatePost = ({ title, content, categories }) => {
  const errors = {};

  // Title: non-empty, min length 3 chars, max 100 chars (example)
  if (!title) {
    errors.title = "Title is required.";
  } else if (title.length < 3) {
    errors.title = "Title must be at least 3 characters.";
  } else if (title.length > 100) {
    errors.title = "Title must be at most 100 characters.";
  }

  // Content: non-empty, min length 10 chars (example)
  if (!content) {
    errors.content = "Content is required.";
  } else if (content.length < 10) {
    errors.content = "Content must be at least 10 characters.";
  }

  // Categories: must be array with 1 to maxCategories items
  if (!Array.isArray(categories) || categories.length === 0) {
    errors.categories = "Select at least one category.";
  } else if (categories.length > maxCategories) {
    errors.categories = `Select no more than ${maxCategories} categories.`;
  }

  return errors;
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
      PopupMessage("Invalid post data. Please check your input.");
      break;
    case 401:
      PopupMessage("Unauthorized. Please sign in.");
      break;
    case 500:
      PopupMessage("Server error. Please try again later.");
      break;
    default:
      PopupMessage("Unexpected error occurred.");
  }
};
