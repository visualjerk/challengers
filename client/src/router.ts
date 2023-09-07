import { createRouter, createWebHistory } from 'vue-router'
import { addAccountRouting } from './account/account-routing'
import { addGameRouting } from './game/game-routing'

const router = createRouter({
  history: createWebHistory(),
  routes: [],
})

addAccountRouting(router)
addGameRouting(router)

export { router }
