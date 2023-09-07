import { RouteRecordRaw, Router, NavigationGuard } from 'vue-router'
import { isAuthenticated } from './account-api'

const createAccountRoute: RouteRecordRaw = {
  name: 'CreateAccount',
  path: '/account/create',
  component: () => import('./create-account.vue'),
}

const accountMiddleware: NavigationGuard = (to) => {
  if (to.name !== createAccountRoute.name && !isAuthenticated()) {
    return {
      name: createAccountRoute.name,
    }
  }
}

export function addAccountRouting(router: Router) {
  router.addRoute(createAccountRoute)
  router.beforeEach(accountMiddleware)
}
