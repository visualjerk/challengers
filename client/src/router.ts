import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import { addAccountRouting } from './account/account-routing'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: Home,
    },
    {
      path: '/game/:id',
      component: () => import('./views/Game.vue'),
    },
  ],
})

addAccountRouting(router)

export { router }
