import { createApp } from "vue";
import "./style.css";
import App from "./App.vue";
import router from "./router";
import { i18n } from "./i18n";
import { useAuth } from "./composables/useAuth";

const app = createApp(App);
app.use(i18n);
app.use(router);

// Initialize auth state on app startup
// This will also set up the unauthorized event listener
const { initAuth } = useAuth();
initAuth();

app.mount(document.body);
