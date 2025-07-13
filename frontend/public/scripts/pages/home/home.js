import { NavBar } from "./nav.js";
import { CategoriesFilter } from "./filter.js";
import { PostsFeed } from "./feed.js";
import { Chat } from "../chat.js";
export { Home };


const Home = {
    html:  NavBar.html +`<div id="layout">${CategoriesFilter.html + PostsFeed.html}</div>${Chat.html}`,
  setup: () => {
    NavBar.setup();
    CategoriesFilter.setup();
    PostsFeed.setup();
    Chat.setup();
  }
};




