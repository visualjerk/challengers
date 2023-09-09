import { RouteRecordRaw, Router, NavigationGuard } from 'vue-router'
import { isAuthenticated } from './account-api'

const createAccountRoute: RouteRecordRaw = {
  name: 'CreateAccount',
  path: '/account/create',
  component: () => import('./create-account.vue'),
}

const authenticationMiddleware: NavigationGuard = async (to) => {
  if (to.name !== createAccountRoute.name && !(await isAuthenticated())) {
    return {
      name: createAccountRoute.name,
      query: {
        redirectTo: to.fullPath,
      },
    }
  }
}

export function addAccountRouting(router: Router) {
  router.addRoute(createAccountRoute)
  router.beforeEach(authenticationMiddleware)
}
