import { RouteRecordRaw, Router } from 'vue-router'

const gameListRoute: RouteRecordRaw = {
  name: 'GameList',
  path: '/',
  component: () => import('./game-list.vue'),
}

const gameRoute: RouteRecordRaw = {
  name: 'Game',
  path: '/game/:id',
  component: () => import('./game.vue'),
}

export function addGameRouting(router: Router) {
  router.addRoute(gameListRoute)
  router.addRoute(gameRoute)
}
