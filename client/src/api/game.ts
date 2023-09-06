import { GameClient } from '../../api/game.client'
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'

const TOKEN_KEY = 'challengers-token'

function getAuthToken() {
  return localStorage.getItem(TOKEN_KEY)
}

function setAuthToken(token: string) {
  return localStorage.setItem(TOKEN_KEY, token)
}

const rpcTransport = new GrpcWebFetchTransport({
  baseUrl: 'http://0.0.0.0:50051',
  interceptors: [
    {
      interceptUnary(next, method, input, options) {
        const token = getAuthToken()

        if (token) {
          options.meta ??= {}
          options.meta.Authorization = token
        }
        return next(method, input, options)
      },
    },
  ],
})
export const gameClient = new GameClient(rpcTransport)

export async function createAccount() {
  const { response } = await gameClient.createAccount({})

  if (response.token) {
    setAuthToken(response.token)
  }
}
