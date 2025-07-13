import { getPayload, loadPosts, state} from "./feed.js";
import { Browse } from "../../router.js";
import { PopupMessage } from "../../tools.js";
export { CategoriesFilter };

const personalCategories = [
  {
    id: "all",
    name: "All Posts",
    description: "Browse all posts",
    icon: "list",
  },
  {
    id: "mine",
    name: "My Posts",
    description: "Your personal posts",
    icon: "person",
  }
]; 

const CategoriesFilter = {
  html: `<div id="categories-container"></div>`,
  setup: () => {
    const categoriesContainer = document.getElementById("categories-container");
    if (!categoriesContainer) return;

    loadCategories({ target: "all" }, categoriesContainer);



  }
};

const loadCategories = async (payload, container) => {
  try {
    const response = await fetch('/categories', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(payload),
    });

    const text = await response.text();

    switch (response.status) {
      case 200: {
        try {
          const categories = JSON.parse(text);
          
          const combinedCategories = [...personalCategories, ...categories];
          combinedCategories.forEach(category => {
            const categoryDiv = document.createElement("div");
            categoryDiv.classList.add("category-card");
            categoryDiv.tabIndex = 0;

            const icon = document.createElement("span");
            icon.classList.add("material-symbols-outlined");
            icon.textContent = category.icon;
            icon.setAttribute('aria-label', category.name);

            const infoDiv = document.createElement("div");
            const name = document.createElement("h4");
            name.textContent = category.name;

            const description = document.createElement("p");
            description.textContent = category.description;

            if (category.id === "all") {
              categoryDiv.classList.add("selected")
              state.currentCategoryId = "all"
            }

            categoryDiv.appendChild(icon);
            infoDiv.appendChild(name);
            infoDiv.appendChild(description);
            categoryDiv.appendChild(infoDiv);
            container.appendChild(categoryDiv);

            categoryDiv.addEventListener("click", () => { 
              if (state.loading) return

              container.querySelectorAll(".category-card.selected").forEach(el => el.classList.remove("selected"));
              categoryDiv.classList.add("selected");
              
              state.currentCategoryId = category.id;

              const postsContainer = document.getElementById("posts-container");
              if (postsContainer) {
                postsContainer.innerHTML = "";
                state.lastPostId = -1;
                state.noMorePosts = false;
                state.loading = false;
                loadPosts(getPayload(), postsContainer);
              }
            });

            categoryDiv.addEventListener("keydown", e => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                categoryDiv.click();
              }
            });
          });
        } catch (err) {
          PopupMessage('Oops, invalid JSON received.')
          
        }
        break;
      }
      case 204: {
        PopupMessage('No categories found','info', 15)
      
        break;
      }
      case 401: {
        localStorage.clear();
        Browse('/signin');
        break;
      }
      default: {
          PopupMessage('Oops, something went wrong')
        }
      }
    } catch (err) {
    PopupMessage('Oops, something went wrong')
  }
};

