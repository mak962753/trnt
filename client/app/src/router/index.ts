import {createRouter, createWebHashHistory, RouteRecordRaw} from "vue-router";
import Home from "@/views/home.vue";

const routes: RouteRecordRaw[] = [
    {path: '/', name: 'Home', component: Home},
];

const router = createRouter({
    history: createWebHashHistory(),
    routes
});

export default router;