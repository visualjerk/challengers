import { API_BASE_URL } from './../config/api-config'
import { AccountClient } from '../../api/account.client'

import { computed, ref, unref } from 'vue'
import { useStorage } from '@vueuse/core'
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'
import { RpcInterceptor } from '@protobuf-ts/runtime-rpc'

const STORAGE_KEY = 'challengers-account'

type AuthStore = {
  token?: string
}

function getAuthToken() {
  const store = localStorage.getItem(STORAGE_KEY)
  if (!store) {
    return null
  }

  return JSON.parse(store)?.token
}

export function isAuthenticated() {
  return !!getAuthToken()
}

export const AuthenticationInterceptor: RpcInterceptor = {
  interceptUnary(next, method, input, options) {
    const token = getAuthToken()

    if (token) {
      options.meta ??= {}
      options.meta.Authorization = token
    }
    return next(method, input, options)
  },
}

export function useAccountApi() {
  const store = useStorage<AuthStore>(STORAGE_KEY, {})
  const isAuthenticated = computed(() => !!unref(store).token)

  const rpcTransport = new GrpcWebFetchTransport({
    baseUrl: API_BASE_URL,
    interceptors: [AuthenticationInterceptor],
  })
  const accountApi = new AccountClient(rpcTransport)

  const pending = ref(false)
  async function create(name: string) {
    pending.value = true

    try {
      const { response } = await accountApi.createAccount({
        name,
      })

      if (response.token) {
        store.value.token = response.token
      }
    } catch (e) {
      console.error(e)
    }
    pending.value = false
  }

  return {
    create,
    pending,
    isAuthenticated,
  }
}
