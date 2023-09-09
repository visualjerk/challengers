import { API_BASE_URL } from './../config/api-config'
import { AccountClient } from '../../api/account.client'

import { ref } from 'vue'
import { useStorage } from '@vueuse/core'
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'
import { RpcInterceptor } from '@protobuf-ts/runtime-rpc'

const STORAGE_KEY = 'challengers-account'

type AuthStore = {
  token?: string
}

function getAuthToken(): null | string {
  const store = localStorage.getItem(STORAGE_KEY)
  if (!store) {
    return null
  }

  return JSON.parse(store)?.token
}

function removeAuthToken() {
  localStorage.removeItem(STORAGE_KEY)
}

export const authenticationInterceptor: RpcInterceptor = {
  interceptUnary(next, method, input, options) {
    const token = getAuthToken()

    if (token) {
      options.meta ??= {}
      options.meta.Authorization = token
    }
    return next(method, input, options)
  },
}

const rpcTransport = new GrpcWebFetchTransport({
  baseUrl: API_BASE_URL,
  interceptors: [authenticationInterceptor],
})
const accountApi = new AccountClient(rpcTransport)

async function verifyAuthToken(token: string): Promise<boolean> {
  try {
    await accountApi.verifyAccount({
      token,
    })
  } catch (e) {
    console.error(e)
    return false
  }
  return true
}

export async function isAuthenticated(): Promise<boolean> {
  const token = getAuthToken()
  if (!token) {
    return false
  }
  const verified = await verifyAuthToken(token)
  if (!verified) {
    removeAuthToken()
  }
  return verified
}

export function useAccountApi() {
  const store = useStorage<AuthStore>(STORAGE_KEY, {})

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
  }
}
