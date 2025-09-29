import { createRouter, createWebHistory } from 'vue-router';
import App from '../App.vue';

const routes = [
    {
        path: '/',
        name: 'Home',
        component: App,
    },
    {
        path: '/invite/:sqid',
        name: 'Invite',
        component: App,
        beforeEnter: (to, from, next) => {
            const joinToken = to.params.sqid;
            localStorage.setItem('rela_join_token', joinToken);
            
            next('/');
        }
    }
];

const router = createRouter({
    history: createWebHistory(),
    routes
});

export default router;