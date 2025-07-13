import { apiRequest, timeAgo, PopupMessage } from "../../tools.js";
import { Browse } from "../../router.js";

export { PostsFeed };
export { loadPosts, getPayload, state };

// State management
const state = {
  lastPostId: -1,
  currentCategoryId: null,
  loading: false,
  noMorePosts: false,
};

// Payload builder based on current state
const getPayload = () => {
  const { currentCategoryId, lastPostId } = state;
  if (!currentCategoryId || currentCategoryId === "all") {
    return { target: "feed", start_id: lastPostId, n_post: 5 };
  } else if (currentCategoryId === "mine") {
    return { target: "user", start_id: lastPostId, n_post: 5 };
  } else {
    return { target: "category", start_id: lastPostId, n_post: 5, category_id: currentCategoryId };
  }
};

// PostsFeed component
const PostsFeed = {
  html: `
    <head>
      <link rel="stylesheet" href="/public/styles/feed.css">
    </head>
    <div id="posts-container"></div>
  `,
  setup: () => {
    const postsContainer = document.getElementById("posts-container");
    if (!postsContainer) return;

    state.lastPostId = -1;
    state.noMorePosts = false;
    state.loading = false;

    loadPosts(getPayload(), postsContainer);
    window.addEventListener("scroll", () => {
      const nearBottom = window.innerHeight + window.scrollY >= document.body.offsetHeight - 300;
      if (nearBottom && !state.loading && !state.noMorePosts) {
        state.loading = true;
        loadPosts(getPayload(), postsContainer).finally(() => {
          setTimeout(() => {
            state.loading = false;
          }, 500);
        });
      }
    });
  },
}

// Load posts asynchronously
const loadPosts = async (payload, postsContainer) => {
  if (state.noMorePosts) return;

  const { status, data, error } = await apiRequest('/posts', payload);

  if (status === 204) {
    if (!state.noMorePosts) {
      PopupMessage('No more posts', 'info', 20);
      state.noMorePosts = true;
    }
    return;
  }

  if (status === 401) {
    localStorage.clear();
    Browse('/signin');
    return;
  }

  if (status >= 400) {
    PopupMessage('Oops, something went wrong');
    return;
  }

  data.forEach(post => createPostElement(post, postsContainer));
  if (data.length > 0) {
    state.lastPostId = data[data.length - 1].id;
  }
};


// Create a post DOM element
const createPostElement = (post, postsContainer) => {
  const article = document.createElement("article");
  article.id = "post";
  article.setAttribute("post-id", post.id);

  // Header
  const postHeader = document.createElement("header");
  postHeader.id = "post-header";

  const img = document.createElement("img");
  img.id = "profile-img";
  img.src = post.user_img || "default-profile.png";
  img.alt = "User Profile";

  const userInfo = document.createElement("div");
  userInfo.id = "user-info";

  const username = document.createElement("span");
  username.className = "username";
  username.textContent = post.username;

  const createdAt = document.createElement("span");
  createdAt.className = "time-ago";
  createdAt.textContent = timeAgo(post.created_at);

  userInfo.appendChild(username);
  userInfo.appendChild(createdAt);

  postHeader.appendChild(img);
  postHeader.appendChild(userInfo);

  // Title
  const title = document.createElement("h2");
  title.id = "post-title";
  title.textContent = post.title;

  // Content
  const content = document.createElement("p");
  content.id = "post-content";
  content.textContent = post.content;

  // Footer
  const footer = document.createElement("div");
  footer.className = "post-footer";

  const categoriesList = document.createElement("ul");
  categoriesList.className = "categories";

  (post.categories || []).slice(0, 3).forEach(category => {
    const li = document.createElement("li");
    li.textContent = category.name;
    categoriesList.appendChild(li);
  });

  const commentIcon = createCommentIcon(post, postsContainer);

  footer.appendChild(categoriesList);
  footer.appendChild(commentIcon);

  // Assemble article
  article.appendChild(postHeader);
  article.appendChild(title);
  article.appendChild(content);
  article.appendChild(footer);

  postsContainer.appendChild(article);
};

// Create comment icon with click handler
const createCommentIcon = (post) => {
  const commentIcon = document.createElement("span");
  commentIcon.classList.add("material-symbols-outlined");
  commentIcon.id = "comment-btn";
  commentIcon.alt = "Comment";
  commentIcon.textContent = "chat";
  commentIcon.title = "View comments";

  commentIcon.addEventListener("click", () => {
    document.body.style.overflow = 'hidden';
    openCommentsPopup(post.id)
  });

  return commentIcon;
};

function prependComment(container, comment) {
  const emptyMsg = container.querySelector(".no-comments-msg");
  if (emptyMsg) emptyMsg.remove();

  const commentDiv = document.createElement("div");
  commentDiv.className = "comment";

  const metaDiv = document.createElement("div");
  metaDiv.textContent = `${comment.username} • just now`;
  metaDiv.innerHTML = `<strong>${localStorage.getItem("username") || "You"}</strong> • Just Now`;

  // metaDiv.textContent = `User: ${comment.user_id} • ${timeAgo(comment.created_at)}`;

  const contentP = document.createElement("p");
  contentP.textContent = comment.content;

  commentDiv.appendChild(metaDiv);
  commentDiv.appendChild(contentP);

  container.insertBefore(commentDiv, container.firstChild);
}

// Render comments popup
const openCommentsPopup = (postID) => {
  let lastCommentId = -1
  let loading = false;
  let noMoreComments = false;

  console.log("Comments rendered!");

  // Create overlayet lastCommentId = -
  const overlay = document.createElement('div');
  overlay.id = 'popup-overlay';
  overlay.onclick = (e) => {
    if (e.target === overlay) {
      document.body.style.overflow = '';
      overlay.remove()
    };
  };
  // Create popup content container
  const popup = document.createElement('div');
  popup.id = 'popup-content';

  // Close button
  const closeBtn = document.createElement('button');
  closeBtn.id = 'popup-close';
  closeBtn.innerHTML = '&times;';
  closeBtn.onclick = () => {
    document.body.style.overflow = '';
    overlay.remove()
  };
  popup.appendChild(closeBtn);

  // Scrollable comment list
  const commentList = document.createElement('div');
  commentList.className = 'comment-list';
  popup.appendChild(commentList);

  const commentInput = document.createElement("input");
  commentInput.className = "comment-input";
  commentInput.name = "type-comment";
  commentInput.type = "text";
  commentInput.placeholder = "Type a comment...";

  commentInput.addEventListener("keydown", async (event) => {
    if (event.key === "Enter") {
      const content = commentInput.value.trim();
      if (!content) {
        PopupMessage("Comment cannot be empty", "error");
        return;
      }

      const payload = {
        post_id: postID,
        content: content,
      };

      const { status, data, error } = await apiRequest("/newcomment", payload, "POST");

      if (status === 200) {
        console.log("Comment submitted", data);
        commentInput.value = "";
        prependComment(commentList, data);
      } else {
        PopupMessage("Couldn't post comment", "error");
        console.error(error);
      }

    }
  });

  popup.appendChild(commentInput);

  // Append and attach
  overlay.appendChild(popup);
  document.body.appendChild(overlay);

  // Infinite scroll inside the popup
  commentList.addEventListener('scroll', async () => {
    const nearBottom = commentList.scrollTop + commentList.clientHeight >= commentList.scrollHeight - 50
    if (nearBottom && !loading && !noMoreComments) {
      loading = true
      await fetchAndRenderComments().finally(() => {
        setTimeout(() => {
          loading = false;
        }, 500);
      })
    }
  });

  fetchAndRenderComments();

  async function fetchAndRenderComments() {
    if (noMoreComments) return;

    loading = true;

    const payload = {
      post_id: postID,
      start_id: lastCommentId,
      n_comment: 5,
    };

    const { status, data, error } = await apiRequest("/comments", payload, "POST");

    if (error || status >= 400) {
      PopupMessage("Oops, something went wrong", "error");
      return;
    }

    if (status === 204 || data === null || (Array.isArray(data) && data.length === 0)) {
      noMoreComments = true;
      if (lastCommentId === -1) {
        const emptyMsg = document.createElement('p');
        emptyMsg.className = 'no-comments-msg'
        emptyMsg.textContent = 'No comments to display.';
        commentList.appendChild(emptyMsg);
      } else {
        PopupMessage('No more comments', 'info');
      }
      return;
    }

    data.forEach((comment) => {
      const commentDiv = document.createElement('div');
      commentDiv.className = 'comment';

      const contentP = document.createElement('p');
      contentP.textContent = comment.content;

      const metaDiv = document.createElement('div');
      metaDiv.innerHTML = `<strong>${comment.username}</strong> • ${timeAgo(comment.created_at)}`;

      commentDiv.appendChild(metaDiv);
      commentDiv.appendChild(contentP);
      commentList.appendChild(commentDiv);
    });

    lastCommentId = data[data.length - 1].id;
    loading = false;
  }
};
