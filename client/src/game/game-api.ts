import { GameClient } from '../../api/game.client'
export { GameEvent, GameState } from '../../api/game'

import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'
import { API_BASE_URL } from '../config/api-config'
import { authenticationInterceptor } from '../account/account-api'

const rpcTransport = new GrpcWebFetchTransport({
  baseUrl: API_BASE_URL,
  interceptors: [authenticationInterceptor],
})
export const gameClient = new GameClient(rpcTransport)
